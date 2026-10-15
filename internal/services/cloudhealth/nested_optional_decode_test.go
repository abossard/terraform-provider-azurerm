// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cloudhealth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	ctyjson "github.com/hashicorp/go-cty/cty/json"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-azurerm/internal/clients"
	"github.com/hashicorp/terraform-provider-azurerm/internal/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
)

func schemaAt(t *testing.T, resource sdk.Resource, path string) *pluginsdk.Schema {
	t.Helper()
	fields := resource.Arguments()
	var field *pluginsdk.Schema
	for key := range strings.SplitSeq(path, ".") {
		field = fields[key]
		if field == nil {
			t.Fatalf("missing schema %s", path)
		}
		if nested, ok := field.Elem.(*pluginsdk.Resource); ok {
			fields = nested.Schema
		}
	}
	return field
}

func TestCloudHealthStringBounds(t *testing.T) {
	type bound struct {
		resource sdk.Resource
		path     string
		max      int
	}
	var bounds []bound
	for _, tc := range lifecycleCases()[1:] {
		bounds = append(bounds, bound{tc.resource, "display_name", 260})
	}
	for _, kind := range []string{"azure_resource", "azure_log_analytics", "azure_monitor_workspace"} {
		prefix := "signal_group." + kind + ".signal."
		fields := map[string]int{"display_name": 260, "time_grain": 100, "data_unit": 100}
		if kind == "azure_resource" {
			fields["metric_namespace"], fields["metric_name"], fields["dimension_filter"] = 256, 256, 2048
		} else {
			fields["query_text"] = 5000
		}
		if kind == "azure_log_analytics" {
			fields["value_column_name"] = 100
		}
		for field, max := range fields {
			bounds = append(bounds, bound{HealthModelEntityResource{}, prefix + field, max})
		}
	}
	for _, kind := range []string{"azure_resource_metric", "log_analytics_query", "prometheus_metrics_query"} {
		fields := map[string]int{"time_grain": 100}
		if kind == "azure_resource_metric" {
			fields["metric_namespace"], fields["metric_name"], fields["dimension_filter"] = 256, 256, 2048
		} else {
			fields["query_text"] = 5000
		}
		if kind == "log_analytics_query" {
			fields["value_column_name"] = 100
		}
		for field, max := range fields {
			bounds = append(bounds, bound{HealthModelSignalDefinitionResource{}, kind + "." + field, max})
		}
	}
	bounds = append(bounds,
		bound{HealthModelEntityResource{}, "icon.name", 100},
		bound{HealthModelEntityResource{}, "icon.custom_data_base64", 5000},
		bound{HealthModelEntityResource{}, "alert.unhealthy.description", 1000},
		bound{HealthModelEntityResource{}, "alert.degraded.description", 1000},
		bound{HealthModelDiscoveryRuleResource{}, "resource_graph_query", 5000},
		bound{HealthModelSignalDefinitionResource{}, "data_unit", 100},
	)
	for _, tc := range bounds {
		t.Run(tc.resource.ResourceType()+"/"+tc.path, func(t *testing.T) {
			field := schemaAt(t, tc.resource, tc.path)
			valid := func(length int) string {
				switch {
				case strings.HasSuffix(tc.path, "time_grain"):
					return "PT" + strings.Repeat("0", length-3) + "S"
				default:
					return strings.Repeat("a", length)
				}
			}
			min := 1
			if strings.HasSuffix(tc.path, "time_grain") || strings.HasSuffix(tc.path, "custom_data_base64") {
				min = 4
			}
			for _, length := range []int{0, min, tc.max, tc.max + 4} {
				value := ""
				if length > 0 {
					value = valid(length)
				}
				_, errors := field.ValidateFunc(value, tc.path)
				expectError := length == 0 || length > tc.max
				if (len(errors) > 0) != expectError {
					t.Fatalf("length %d: expected error=%t, got %v", length, expectError, errors)
				}
			}
			if strings.HasSuffix(tc.path, "time_grain") || strings.HasSuffix(tc.path, "custom_data_base64") {
				if _, errors := field.ValidateFunc("not valid", tc.path); len(errors) == 0 {
					t.Fatal("syntax guard lost")
				}
			}
		})
	}
}

func TestCloudHealthManagedIdentity(t *testing.T) {
	id := "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/Group/providers/Microsoft.ManagedIdentity/userAssignedIdentities/Identity"
	for _, tc := range []struct {
		value string
		valid bool
	}{
		{"SystemAssigned", true},
		{id, true},
		{id + strings.Repeat("a", 500-len(id)), true},
		{"", false},
		{"identity-name", false},
		{regressionModelID, false},
		{id + "/extra", false},
		{id + strings.Repeat("a", 501-len(id)), false},
	} {
		t.Run(tc.value, func(t *testing.T) {
			_, errors := HealthModelAuthenticationSettingResource{}.Arguments()["managed_identity"].ValidateFunc(tc.value, "managed_identity")
			if (len(errors) == 0) != tc.valid {
				t.Fatalf("expected valid=%t, got %v", tc.valid, errors)
			}
		})
	}
}

func rawConfigValue(t *testing.T, config map[string]any) cty.Value {
	t.Helper()
	encoded, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	var raw ctyjson.SimpleJSONValue
	if err := json.Unmarshal(encoded, &raw); err != nil {
		t.Fatal(err)
	}
	return raw.Value
}

func planConfig(t *testing.T, resource sdk.Resource, config map[string]any, raw cty.Value) error {
	t.Helper()
	wrapped := sdk.WrappedResource(resource)
	if diagnostics := wrapped.Validate(terraform.NewResourceConfigRaw(config)); diagnostics.HasError() {
		return fmt.Errorf("schema validation: %v", diagnostics)
	}
	_, err := wrapped.Diff(context.Background(), &terraform.InstanceState{RawConfig: raw}, terraform.NewResourceConfigRaw(config), &clients.Client{})
	return err
}

func signalConfig(kind string, signal map[string]any) map[string]any {
	group := map[string]any{"authentication_setting_name": "auth", "signal": []any{signal}}
	switch kind {
	case "azure_resource":
		group["target_resource_id"] = regressionModelID
	case "azure_log_analytics":
		group["log_analytics_workspace_id"] = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/Group/providers/Microsoft.OperationalInsights/workspaces/Workspace"
	case "azure_monitor_workspace":
		group["monitor_workspace_id"] = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/Group/providers/Microsoft.Monitor/accounts/Workspace"
	}
	return map[string]any{"name": "Child", "health_model_id": regressionModelID, "signal_group": []any{map[string]any{kind: []any{group}}}}
}

func inlineSignal(kind string) map[string]any {
	result := map[string]any{"name": "signal", "evaluation_rule": []any{map[string]any{"unhealthy": []any{map[string]any{"operator": "GreaterThan", "threshold": float64(0)}}}}}
	if kind == "azure_resource" {
		result["metric_namespace"], result["metric_name"], result["aggregation_type"], result["time_grain"] = "namespace", "metric", "Average", "PT1M"
	} else {
		result["query_text"] = "query"
	}
	return result
}

func TestCloudHealthPlanSignalInstances(t *testing.T) {
	for _, kind := range []string{"azure_resource", "azure_log_analytics", "azure_monitor_workspace"} {
		for _, mode := range []string{"template", "inline", "missing evaluation_rule", "missing query_text", "missing metric_namespace", "missing metric_name", "missing aggregation_type", "missing time_grain"} {
			if strings.Contains(mode, "missing") && kind == "azure_resource" && mode == "missing query_text" {
				continue
			}
			if strings.Contains(mode, "missing") && kind != "azure_resource" && mode != "missing query_text" && mode != "missing evaluation_rule" {
				continue
			}
			t.Run(kind+"/"+mode, func(t *testing.T) {
				signal := inlineSignal(kind)
				expectedError := ""
				switch {
				case mode == "template":
					signal = map[string]any{"name": "signal", "signal_definition_name": "definition"}
				case strings.HasPrefix(mode, "missing "):
					expectedError = strings.TrimPrefix(mode, "missing ")
					delete(signal, expectedError)
				}
				config := signalConfig(kind, signal)
				err := planConfig(t, HealthModelEntityResource{}, config, rawConfigValue(t, config))
				if expectedError == "" && err != nil {
					t.Fatal(err)
				}
				if expectedError != "" && (err == nil || !strings.Contains(err.Error(), expectedError)) {
					t.Fatalf("expected %s diagnostic, got %v", expectedError, err)
				}
			})
		}
	}
}

func TestCloudHealthPlanOptionalBlocks(t *testing.T) {
	for _, block := range []string{"signal_group", "alert"} {
		for _, mode := range []string{"omitted", "empty", "first", "second", "both"} {
			t.Run(block+"/"+mode, func(t *testing.T) {
				config := map[string]any{"name": "Child", "health_model_id": regressionModelID, "tags": map[string]any{"signal_group": "tag", "alert": "tag", "signal": "tag"}}
				children := map[string]any{}
				if mode == "first" || mode == "both" {
					if block == "alert" {
						children["unhealthy"] = []any{map[string]any{"severity": "Sev0"}}
					} else {
						children["dependency"] = []any{map[string]any{"aggregation_type": "WorstOf"}}
					}
				}
				if mode == "second" || mode == "both" {
					if block == "alert" {
						children["degraded"] = []any{map[string]any{"severity": "Sev1"}}
					} else {
						children["azure_resource"] = []any{map[string]any{"authentication_setting_name": "auth", "target_resource_id": regressionModelID}}
					}
				}
				if mode != "omitted" {
					config[block] = []any{children}
				}
				err := planConfig(t, HealthModelEntityResource{}, config, rawConfigValue(t, config))
				if mode == "empty" {
					if err == nil || !strings.Contains(err.Error(), block) {
						t.Fatalf("expected %s diagnostic, got %v", block, err)
					}
				} else if err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestCloudHealthDefinitionKinds(t *testing.T) {
	fields := []string{"azure_resource_metric", "log_analytics_query", "prometheus_metrics_query"}
	for _, count := range []int{0, 1, 2, 3} {
		t.Run(fmt.Sprintf("kind count %d", count), func(t *testing.T) {
			tc := lifecycleCases()[5]
			config := copyConfig(tc.config)
			delete(config, "azure_resource_metric")
			for _, field := range fields[:count] {
				if field == "azure_resource_metric" {
					config[field] = tc.config[field]
				} else {
					config[field] = []any{map[string]any{"query_text": "query"}}
				}
			}
			err := planConfig(t, tc.resource, config, rawConfigValue(t, config))
			if (err == nil) != (count == 1) {
				t.Fatalf("kind count %d: %v", count, err)
			}
		})
	}
	for _, field := range fields[1:] {
		t.Run(field, func(t *testing.T) {
			tc := lifecycleCases()[5]
			config := copyConfig(tc.config)
			delete(config, "azure_resource_metric")
			config[field] = []any{map[string]any{"query_text": "query"}}
			if err := planConfig(t, tc.resource, config, rawConfigValue(t, config)); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCloudHealthUnknownPlanValues(t *testing.T) {
	for _, kind := range []string{"azure_resource", "azure_log_analytics", "azure_monitor_workspace"} {
		t.Run(kind, func(t *testing.T) {
			config := signalConfig(kind, map[string]any{"name": "signal", "signal_definition_name": "definition"})
			raw := cty.ObjectVal(map[string]cty.Value{
				"signal_group": cty.TupleVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
					kind: cty.TupleVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
						"signal": cty.TupleVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{"name": cty.StringVal("signal"), "signal_definition_name": cty.UnknownVal(cty.String)})}),
					})}),
				})}),
			})
			if err := planConfig(t, HealthModelEntityResource{}, config, raw); err != nil {
				t.Fatal(err)
			}
		})
	}
	for _, resource := range []sdk.Resource{HealthModelEntityResource{}, HealthModelSignalDefinitionResource{}} {
		for _, state := range []string{"unhealthy", "degraded"} {
			for _, operatorUnknown := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/operator unknown=%t", resource.ResourceType(), state, operatorUnknown), func(t *testing.T) {
					rules := map[string]any{"unhealthy": []any{map[string]any{"operator": "Dynamic"}}}
					rules[state] = []any{map[string]any{"operator": "GreaterThan", "threshold": float64(1)}}
					config := copyConfig(lifecycleCases()[5].config)
					config["evaluation_rule"] = []any{rules}
					if _, entity := resource.(HealthModelEntityResource); entity {
						signal := inlineSignal("azure_resource")
						signal["evaluation_rule"] = []any{rules}
						config = signalConfig("azure_resource", signal)
					}
					operator, threshold := cty.StringVal("GreaterThan"), cty.UnknownVal(cty.Number)
					if operatorUnknown {
						operator, threshold = cty.UnknownVal(cty.String), cty.NullVal(cty.Number)
					}
					raw, err := cty.Transform(rawConfigValue(t, config), func(path cty.Path, value cty.Value) (cty.Value, error) {
						inState := false
						for _, step := range path {
							if attr, ok := step.(cty.GetAttrStep); ok && attr.Name == state {
								inState = true
							}
						}
						if inState {
							if attr, ok := path[len(path)-1].(cty.GetAttrStep); ok {
								switch attr.Name {
								case "operator":
									return operator, nil
								case "threshold":
									return threshold, nil
								}
							}
						}
						return value, nil
					})
					if err != nil {
						t.Fatal(err)
					}
					if err := planConfig(t, resource, config, raw); err != nil {
						t.Fatal(err)
					}
				})
			}
		}
	}
}

func TestCloudHealthPlanStaticThresholds(t *testing.T) {
	for _, owner := range []string{"entity", "definition"} {
		for _, state := range []string{"unhealthy", "degraded"} {
			for _, value := range []any{nil, float64(0), float64(12)} {
				t.Run(fmt.Sprintf("%s/%s/%v", owner, state, value), func(t *testing.T) {
					threshold := map[string]any{"operator": "GreaterThan"}
					if value != nil {
						threshold["threshold"] = value
					}
					rules := map[string]any{"unhealthy": []any{map[string]any{"operator": "Dynamic"}}}
					rules[state] = []any{threshold}
					var resource sdk.Resource = HealthModelEntityResource{}
					config := signalConfig("azure_resource", inlineSignal("azure_resource"))
					signal := config["signal_group"].([]any)[0].(map[string]any)["azure_resource"].([]any)[0].(map[string]any)["signal"].([]any)[0].(map[string]any)
					signal["evaluation_rule"] = []any{rules}
					if owner == "definition" {
						tc := lifecycleCases()[5]
						resource, config = tc.resource, copyConfig(tc.config)
						config["evaluation_rule"] = []any{rules}
					}
					err := planConfig(t, resource, config, rawConfigValue(t, config))
					if value == nil {
						if err == nil || !strings.Contains(err.Error(), "threshold") {
							t.Fatalf("expected missing threshold diagnostic, got %v", err)
						}
					} else if err != nil {
						t.Fatal(err)
					}
				})
			}
		}
	}
}

func TestCloudHealthTemplateLifecycle(t *testing.T) {
	for _, kind := range []string{"azure_resource", "azure_log_analytics", "azure_monitor_workspace"} {
		for _, mode := range []string{"template", "refresh override", "query or metric override", "evaluation override", "inline default"} {
			t.Run(kind+"/"+mode, func(t *testing.T) {
				signal := map[string]any{"name": "signal", "signal_definition_name": "definition"}
				switch mode {
				case "refresh override":
					signal["refresh_interval"] = "PT1M"
				case "query or metric override":
					if kind == "azure_resource" {
						signal["metric_name"] = "metric"
					} else {
						signal["query_text"] = "query"
					}
				case "evaluation override":
					signal["evaluation_rule"] = inlineSignal(kind)["evaluation_rule"]
				case "inline default":
					signal = inlineSignal(kind)
				}
				config := signalConfig(kind, signal)
				raw := rawConfigValue(t, config)
				if err := planConfig(t, HealthModelEntityResource{}, config, raw); err != nil {
					t.Fatal(err)
				}
				var payload map[string]any
				client := lifecycleClient(t, func(w http.ResponseWriter, request *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					if request.Method == http.MethodGet {
						response, err := json.Marshal(payload)
						if err != nil {
							t.Error(err)
						}
						_, _ = w.Write(response)
						return
					}
					if request.Method != http.MethodPut {
						t.Errorf("unexpected method %s", request.Method)
					}
					if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
						t.Error(err)
					}
					_, _ = io.WriteString(w, `{}`)
				})
				client.Features.SkipImportCheckOnCreateAndAllowOverwritingExistingResources = true
				tc := lifecycleCase{resource: HealthModelEntityResource{}, config: config, id: regressionModelID + "/entities/Child"}
				metadata := lifecycleMetadata(t, tc, config, raw, client)
				var decoded HealthModelEntityResourceModel
				if err := metadata.Decode(&decoded); err != nil {
					t.Fatal(err)
				}
				ctx := lifecycleContext(t)
				if mode == "template" {
					definition := lifecycleCases()[5]
					definition.config = copyConfig(definition.config)
					definition.config["name"] = "definition"
					definition.id = regressionModelID + "/signaldefinitions/definition"
					definition.config["refresh_interval"] = "PT5M"
					if kind != "azure_resource" {
						delete(definition.config, "azure_resource_metric")
						definitionKind := "log_analytics_query"
						if kind == "azure_monitor_workspace" {
							definitionKind = "prometheus_metrics_query"
						}
						definition.config[definitionKind] = []any{map[string]any{"query_text": "query"}}
					}
					definitionRaw := rawConfigValue(t, definition.config)
					if err := planConfig(t, definition.resource, definition.config, definitionRaw); err != nil {
						t.Fatal(err)
					}
					definitionMetadata := lifecycleMetadata(t, definition, definition.config, definitionRaw, client)
					if err := definition.resource.Create().Func(ctx, definitionMetadata); err != nil {
						t.Fatal(err)
					}
					if payload["properties"].(map[string]any)["refreshInterval"] != "PT5M" {
						t.Fatalf("definition interval missing: %#v", payload)
					}
				}
				if err := tc.resource.Create().Func(ctx, metadata); err != nil {
					t.Fatal(err)
				}
				properties := payload["properties"].(map[string]any)
				wireName := map[string]string{"azure_resource": "azureResource", "azure_log_analytics": "azureLogAnalytics", "azure_monitor_workspace": "azureMonitorWorkspace"}[kind]
				wire := properties["signalGroups"].(map[string]any)[wireName].(map[string]any)["signals"].([]any)[0].(map[string]any)
				if mode != "inline default" && wire["signalDefinitionName"] != "definition" {
					t.Fatalf("reference missing: %#v", wire)
				}
				if wire["name"] != "signal" {
					t.Fatalf("name missing: %#v", wire)
				}
				for _, key := range []string{"metricNamespace", "metricName", "aggregationType", "timeGrain", "queryText", "evaluationRules", "refreshInterval"} {
					_, present := wire[key]
					expected := mode == "inline default" && (kind == "azure_resource" && key != "queryText" || kind != "azure_resource" && (key == "queryText" || key == "evaluationRules" || key == "refreshInterval"))
					expected = expected || mode == "refresh override" && key == "refreshInterval" || mode == "evaluation override" && key == "evaluationRules" || mode == "query or metric override" && (kind == "azure_resource" && key == "metricName" || kind != "azure_resource" && key == "queryText")
					if present != expected {
						t.Fatalf("unexpected %s presence=%t expected=%t in %#v", key, present, expected, wire)
					}
				}
				if mode == "template" {
					wire["refreshInterval"] = "PT5M"
				}
				if err := tc.resource.Read().Func(ctx, metadata); err != nil {
					t.Fatal(err)
				}
				var state HealthModelEntityResourceModel
				if err := metadata.Decode(&state); err != nil {
					t.Fatal(err)
				}
				if len(state.SignalGroup) != 1 {
					t.Fatalf("lost signal group: %#v", state)
				}
				prefix := "signal_group.0." + kind + ".0.signal.0."
				if metadata.ResourceData.Get(prefix+"name") != "signal" || mode != "inline default" && metadata.ResourceData.Get(prefix+"signal_definition_name") != "definition" {
					t.Fatalf("name or reference lost in Read state: %#v", state)
				}
				if mode == "template" && metadata.ResourceData.Get(prefix+"refresh_interval") != "" {
					t.Fatalf("inherited interval encoded as an explicit override: %#v", state)
				}
				if mode == "refresh override" {
					planned := signalConfig(kind, map[string]any{"name": "signal", "signal_definition_name": "definition"})
					propsJSON, err := json.Marshal(properties)
					if err != nil {
						t.Fatal(err)
					}
					next := captureEntityUpdate(t, config, planned, rawConfigValue(t, planned), string(propsJSON))
					signal := next["signalGroups"].(map[string]any)[wireName].(map[string]any)["signals"].([]any)[0].(map[string]any)
					if _, present := signal["refreshInterval"]; present {
						t.Fatalf("removed refresh override restored: %#v", signal)
					}
				}
			})
		}
	}
}

// TestDecodeDistinguishesAbsenceOnlyAtTheTopLevel pins the decoder behaviour that the optional
// numeric arguments in this package are built on, and that a reader would otherwise have to infer
// from `internal/sdk/resource_decode.go`.
//
// `Decode` skips a field whose `tfschema` path reports no value, so a `*float64` left out of the
// configuration stays nil - but that check is applied to top-level arguments only. Inside a block
// the decoder walks the already-flattened map and assigns every key it finds, so an omitted
// `TypeFloat` arrives as a pointer to `0` rather than as nil.
//
// The consequence is that a pointer distinguishes "unset" from "zero" at the top level and does not
// distinguish anything at all inside a block, which is why `health_objective_percentage` is modelled
// as `*float64` while the thresholds inside `dependency` are plain `float64` read through the raw
// configuration. Reversing either choice reintroduces a shipped bug: as a pointer at depth,
// `degraded_threshold` sent `0` on every dependency block and the service rejected it.
func TestDecodeDistinguishesAbsenceOnlyAtTheTopLevel(t *testing.T) {
	metadata := sdk.NewResourceMetaData(nil, HealthModelEntityResource{})

	if err := metadata.ResourceData.Set("name", "example"); err != nil {
		t.Fatalf("seeding the resource data: %+v", err)
	}

	// `signal_group.dependency` is the block that carries the thresholds, and the schema writer
	// fills every key the block declares, so omitting `degraded_threshold` here is exactly what a
	// configuration that never mentions it produces.
	if err := metadata.ResourceData.Set("signal_group", []interface{}{
		map[string]interface{}{
			"dependency": []interface{}{
				map[string]interface{}{
					"aggregation_type": "WorstOf",
				},
			},
		},
	}); err != nil {
		t.Fatalf("seeding the resource data: %+v", err)
	}

	var decoded HealthModelEntityResourceModel
	if err := metadata.Decode(&decoded); err != nil {
		t.Fatalf("decoding: %+v", err)
	}

	t.Run("an omitted top-level optional stays nil", func(t *testing.T) {
		if decoded.HealthObjectivePercentage != nil {
			t.Fatalf("expected `health_objective_percentage` to decode as nil when absent but got %v - if this now decodes as a pointer to 0 the top-level pointer no longer distinguishes unset from zero and `health_objective_percentage = 0` cannot be expressed", *decoded.HealthObjectivePercentage)
		}
	})

	t.Run("an omitted nested optional decodes as its zero value", func(t *testing.T) {
		if len(decoded.SignalGroup) != 1 || len(decoded.SignalGroup[0].Dependency) != 1 {
			t.Fatalf("expected one `dependency` block but got %+v", decoded.SignalGroup)
		}

		if got := decoded.SignalGroup[0].Dependency[0].DegradedThreshold; got != 0 {
			t.Fatalf("expected an omitted nested `degraded_threshold` to decode as 0 but got %v", got)
		}
	})

	// The two subtests above describe the decoder; this one enforces the conclusion drawn from it.
	// Without it, changing a nested threshold to `*float64` and guarding on `!= nil` compiles, keeps
	// every other test in this package green, and silently sends `0` to the service on every block.
	t.Run("no nested numeric argument is modelled as a pointer", func(t *testing.T) {
		checked := 0
		for _, model := range []interface{}{
			SignalGroupModel{},
			DependencySignalGroupModel{},
			AzureResourceSignalGroupModel{},
			AzureResourceSignalModel{},
			LogAnalyticsSignalGroupModel{},
			LogAnalyticsSignalModel{},
			AzureMonitorWorkspaceSignalGroupModel{},
			PrometheusMetricsSignalModel{},
			EvaluationRuleModel{},
			ThresholdRuleModel{},
			EntityAlertModel{},
			AlertConfigurationModel{},
			CanvasPositionModel{},
			IconModel{},
		} {
			objType := reflect.TypeOf(model)
			for i := 0; i < objType.NumField(); i++ {
				field := objType.Field(i)
				tag, ok := field.Tag.Lookup("tfschema")
				if !ok {
					continue
				}

				checked++
				if field.Type.Kind() != reflect.Ptr {
					continue
				}

				if elem := field.Type.Elem().Kind(); elem == reflect.Float64 || elem == reflect.Int64 || elem == reflect.Bool || elem == reflect.String {
					t.Fatalf("`%s.%s` (%q) is a pointer inside a block, where the decoder cannot report absence - a pointer here is always non-nil and guarding on it sends the zero value to the service. Read configured-ness from the raw configuration instead, as `dependencyDegradedThresholdConfigured` does", objType.Name(), field.Name, tag)
				}
			}
		}

		if checked == 0 {
			t.Fatal("expected to inspect the nested models but found no `tfschema` fields, so this assertion proves nothing")
		}

		t.Logf("inspected %d nested `tfschema` fields", checked)
	})
}
