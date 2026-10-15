// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cloudhealth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-azure-sdk/sdk/auth"
	"github.com/hashicorp/go-azure-sdk/sdk/client/resourcemanager"
	"github.com/hashicorp/go-azure-sdk/sdk/environments"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-azurerm/internal/clients"
	"github.com/hashicorp/terraform-provider-azurerm/internal/common"
	"github.com/hashicorp/terraform-provider-azurerm/internal/sdk"
	cloudhealthclient "github.com/hashicorp/terraform-provider-azurerm/internal/services/cloudhealth/client"
)

const regressionModelID = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/Group/providers/Microsoft.CloudHealth/healthmodels/Model"

type lifecycleCase struct {
	resource sdk.ResourceWithUpdate
	config   map[string]any
	id       string
	response string
}

func lifecycleCases() []lifecycleCase {
	child := func(extra map[string]any) map[string]any {
		config := map[string]any{"name": "Child", "health_model_id": regressionModelID}
		maps.Copy(config, extra)
		return config
	}
	return []lifecycleCase{
		{
			resource: HealthModelResource{},
			config:   map[string]any{"name": "Model", "resource_group_name": "Group", "location": "westus"},
			id:       regressionModelID,
			response: `{"location":"westus","identity":{"type":"SystemAssigned","principalId":"server-only","tenantId":"server-only"},"tags":{"remote":"keep"},"properties":{"provisioningState":"Succeeded"},"id":"server-only","name":"server-only","type":"server-only","systemData":{}}`,
		},
		{
			resource: HealthModelAuthenticationSettingResource{},
			config:   child(map[string]any{"managed_identity": "SystemAssigned"}),
			id:       regressionModelID + "/authenticationsettings/Child",
			response: `{"properties":{"authenticationKind":"ManagedIdentity","managedIdentityName":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/Group/providers/Microsoft.ManagedIdentity/userAssignedIdentities/RemoteIdentity","displayName":"remote","provisioningState":"Succeeded"}}`,
		},
		{
			resource: HealthModelDiscoveryRuleResource{},
			config:   child(map[string]any{"authentication_setting_name": "auth", "resource_graph_query": "resources"}),
			id:       regressionModelID + "/discoveryrules/Child",
			response: `{"properties":{"authenticationSetting":"auth","displayName":"remote","discoverRelationships":"Disabled","addRecommendedSignals":"Disabled","addResourceHealthSignal":"Disabled","specification":{"kind":"ResourceGraphQuery","resourceGraphQuery":"remote query"},"entityName":"server-only","provisioningState":"Succeeded","error":{}}}`,
		},
		{
			resource: HealthModelEntityResource{},
			config:   child(nil),
			id:       regressionModelID + "/entities/Child",
			response: `{"properties":{"displayName":"remote","tags":{"remote":"keep"},"healthObjective":0,"impact":"Limited","provisioningState":"Succeeded","discoveredBy":"server-only","healthState":"Healthy","signalAggregationGroups":[{"name":"rollup","members":["cpu"],"aggregationType":"WorstOf","aggregatedHealthState":"Healthy","unresolvedMembers":["server-only"]}],"signalGroups":{"azureResource":{"authenticationSetting":"remote auth","azureResourceId":"remote ID","azureResourceKind":"remote kind","resourceHealth":{"enabled":"Enabled","signalName":"server-only","status":{}},"signals":[{"name":"metric","signalKind":"AzureResourceMetric","metricName":"remote metric","status":{}}]},"azureLogAnalytics":{"authenticationSetting":"remote auth","logAnalyticsWorkspaceResourceId":"remote ID","signals":[{"name":"log","signalKind":"LogAnalyticsQuery","queryText":"remote query","status":{}}]},"azureMonitorWorkspace":{"authenticationSetting":"remote auth","azureMonitorWorkspaceResourceId":"remote ID","signals":[{"name":"prom","signalKind":"PrometheusMetricsQuery","queryText":"remote query","status":{}}]},"external":{"signals":[]}}}}`,
		},
		{
			resource: HealthModelRelationshipResource{},
			config:   child(map[string]any{"parent_entity_name": "parent", "child_entity_name": "child"}),
			id:       regressionModelID + "/relationships/Child",
			response: `{"properties":{"parentEntityName":"remote parent","childEntityName":"remote child","displayName":"remote","tags":{"remote":"keep"},"provisioningState":"Succeeded","discoveredBy":"server-only"}}`,
		},
		{
			resource: HealthModelSignalDefinitionResource{},
			config: child(map[string]any{
				"azure_resource_metric": []any{map[string]any{"metric_namespace": "namespace", "metric_name": "metric", "aggregation_type": "Average", "time_grain": "PT1M"}},
				"evaluation_rule":       []any{map[string]any{"unhealthy": []any{map[string]any{"operator": "GreaterThan", "threshold": float64(1)}}}},
			}),
			id:       regressionModelID + "/signaldefinitions/Child",
			response: `{"properties":{"signalKind":"AzureResourceMetric","metricNamespace":"remote namespace","metricName":"remote metric","aggregationType":"Maximum","timeGrain":"PT5M","dimensionFilter":"remote filter","displayName":"remote","dataUnit":"remote unit","refreshInterval":"PT5M","tags":{"remote":"keep"},"evaluationRules":{"unhealthyRule":{"operator":"GreaterThan","threshold":3}},"provisioningState":"Succeeded"}}`,
		},
	}
}

type loopbackTransport struct {
	base *url.URL
}

func (transport loopbackTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	request = request.Clone(request.Context())
	request.URL.Scheme = transport.base.Scheme
	request.URL.Host = transport.base.Host
	return http.DefaultTransport.RoundTrip(request)
}

func lifecycleClient(t *testing.T, handler http.HandlerFunc) *clients.Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/poll" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"status":"Succeeded"}`)
			return
		}
		if request.Method == http.MethodPut || request.Method == http.MethodPatch {
			w.Header().Set("Azure-AsyncOperation", "https://management.azure.com/poll")
			w.Header()["X-Go-Azure-Sdk-Skip-Polling-Delay"] = []string{"true"}
		}
		handler(w, request)
	}))
	t.Cleanup(server.Close)
	base, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	service, err := cloudhealthclient.NewClient(&common.ClientOptions{
		Authorizers: &common.Authorizers{}, Environment: *environments.AzurePublic(),
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, client := range []*resourcemanager.Client{
		service.HealthModels.Client, service.AuthenticationSettings.Client, service.DiscoveryRules.Client,
		service.Entities.Client, service.Relationships.Client, service.SignalDefinitions.Client,
	} {
		client.SetTransport(loopbackTransport{base: base})
		client.AuthorizeRequest = func(context.Context, *http.Request, auth.Authorizer) error { return nil }
		client.DisableRetries = true
	}
	return &clients.Client{CloudHealth: service, Account: &clients.ResourceManagerAccount{SubscriptionId: "00000000-0000-0000-0000-000000000000"}}
}

func lifecycleMetadata(t *testing.T, tc lifecycleCase, planned map[string]any, raw cty.Value, client *clients.Client) sdk.ResourceMetaData {
	t.Helper()
	metadata := sdk.NewResourceMetaData(client, tc.resource)
	wrapped := sdk.WrappedResource(tc.resource)
	initialState := &terraform.InstanceState{RawConfig: raw}
	initialDiff, err := wrapped.Diff(context.Background(), initialState, terraform.NewResourceConfigRaw(tc.config), client)
	if err != nil {
		t.Fatal(err)
	}
	metadata.ResourceData, err = schema.InternalMap(wrapped.Schema).Data(initialState, initialDiff)
	if err != nil {
		t.Fatal(err)
	}
	metadata.ResourceData.SetId(tc.id)
	state := metadata.ResourceData.State()
	state.RawConfig = raw
	state.RawState = rawConfigValue(t, tc.config)
	state.RawPlan = rawConfigValue(t, planned)
	diff, err := wrapped.Diff(context.Background(), state, terraform.NewResourceConfigRaw(planned), client)
	if err != nil {
		t.Fatal(err)
	}
	if diff == nil {
		diff = terraform.NewInstanceDiff()
	}
	diff.RawConfig = raw
	metadata.ResourceData, err = schema.InternalMap(wrapped.Schema).Data(state, diff)
	if err != nil {
		t.Fatal(err)
	}
	return metadata
}

func copyConfig(input map[string]any) map[string]any {
	result := make(map[string]any, len(input))
	maps.Copy(result, input)
	return result
}

func TestCloudHealthUpdatePreservesWritableProperties(t *testing.T) {
	for _, tc := range lifecycleCases() {
		for _, failure := range []string{"", "get error", "nil properties"} {
			t.Run(tc.resource.ResourceType()+"/"+failure, func(t *testing.T) {
				var methods []string
				var payload map[string]any
				client := lifecycleClient(t, func(w http.ResponseWriter, request *http.Request) {
					methods = append(methods, request.Method)
					w.Header().Set("Content-Type", "application/json")
					switch request.Method {
					case http.MethodGet:
						switch failure {
						case "get error":
							w.WriteHeader(http.StatusBadRequest)
							_, _ = io.WriteString(w, `{"error":{"code":"BadRequest","message":"get failed"}}`)
						case "nil properties":
							_, _ = io.WriteString(w, `{}`)
						default:
							_, _ = io.WriteString(w, tc.response)
						}
					case http.MethodPut, http.MethodPatch:
						if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
							t.Error(err)
						}
						_, _ = io.WriteString(w, `{"properties":{"provisioningState":"Succeeded"}}`)
					default:
						t.Errorf("unexpected method %s", request.Method)
					}
				})
				planned := copyConfig(tc.config)
				changedField := "display_name"
				if _, ok := tc.resource.(HealthModelResource); ok {
					changedField = "tags"
					planned[changedField] = map[string]any{"changed": "value"}
				} else {
					planned[changedField] = "changed"
				}
				metadata := lifecycleMetadata(t, tc, planned, cty.NullVal(cty.DynamicPseudoType), client)
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				err := tc.resource.Update().Func(ctx, metadata)
				if failure != "" {
					if err == nil || !reflect.DeepEqual(methods, []string{"GET"}) {
						t.Fatalf("expected retrieval error before PUT, got error=%v methods=%v", err, methods)
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(methods, []string{"GET", "PUT"}) {
					t.Fatalf("expected GET/PUT, got %v", methods)
				}
				var expected map[string]any
				if err := json.Unmarshal([]byte(tc.response), &expected); err != nil {
					t.Fatal(err)
				}
				for _, key := range []string{"id", "name", "type", "systemData"} {
					delete(expected, key)
				}
				properties := expected["properties"].(map[string]any)
				for _, key := range []string{"provisioningState", "discoveredBy", "healthState", "entityName", "error"} {
					delete(properties, key)
				}
				if groups, ok := properties["signalGroups"].(map[string]any); ok {
					delete(groups, "external")
					for _, value := range groups {
						group := value.(map[string]any)
						if health, ok := group["resourceHealth"].(map[string]any); ok {
							delete(health, "signalName")
							delete(health, "status")
						}
						for _, signal := range group["signals"].([]any) {
							delete(signal.(map[string]any), "status")
						}
					}
				}
				if groups, ok := properties["signalAggregationGroups"].([]any); ok {
					for _, group := range groups {
						delete(group.(map[string]any), "aggregatedHealthState")
						delete(group.(map[string]any), "unresolvedMembers")
					}
				}
				if identity, ok := expected["identity"].(map[string]any); ok {
					delete(identity, "principalId")
					delete(identity, "tenantId")
					identity["userAssignedIdentities"] = nil
				}
				if changedField == "tags" {
					expected["tags"] = map[string]any{"changed": "value"}
				} else {
					properties["displayName"] = "changed"
				}
				if !reflect.DeepEqual(payload, expected) {
					t.Fatalf("unexpected PUT:\nactual: %#v\nexpected: %#v", payload, expected)
				}
			})
		}
	}
}

func TestCloudHealthCreateOverwrite(t *testing.T) {
	for _, tc := range lifecycleCases() {
		for _, status := range []int{http.StatusOK, http.StatusNotFound, http.StatusBadRequest} {
			for _, overwrite := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%d/overwrite=%t", tc.resource.ResourceType(), status, overwrite), func(t *testing.T) {
					var methods []string
					client := lifecycleClient(t, func(w http.ResponseWriter, request *http.Request) {
						methods = append(methods, request.Method)
						w.Header().Set("Content-Type", "application/json")
						if request.Method == http.MethodGet {
							w.WriteHeader(status)
							if status != http.StatusOK {
								_, _ = io.WriteString(w, `{"error":{"code":"NotFound","message":"preflight"}}`)
								return
							}

							_, _ = io.WriteString(w, tc.response)
							return
						}
						if request.Method != http.MethodPut {
							t.Errorf("unexpected method %s", request.Method)
						}
						_, _ = io.WriteString(w, `{"properties":{"provisioningState":"Succeeded"}}`)
					})
					client.Features.SkipImportCheckOnCreateAndAllowOverwritingExistingResources = overwrite
					metadata := lifecycleMetadata(t, tc, tc.config, rawConfigValue(t, tc.config), client)
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					err := tc.resource.Create().Func(ctx, metadata)
					allowed := overwrite || status == http.StatusNotFound
					if allowed && err != nil {
						t.Fatal(err)
					}
					if !allowed && err == nil {
						t.Fatal("expected preflight failure")
					}
					if !allowed && status == http.StatusOK && !strings.Contains(err.Error(), "import") {
						t.Fatal(err)
					}
					expected := []string{"GET"}
					if overwrite {
						expected = []string{"PUT"}
					} else if allowed {
						expected = append(expected, "PUT")
					}
					if !reflect.DeepEqual(methods, expected) {
						t.Fatalf("got %v, expected %v", methods, expected)
					}
				})
			}
		}
	}
}

func TestCloudHealthUpdateRefreshPresence(t *testing.T) {
	for _, kind := range []string{"azure_resource", "azure_log_analytics", "azure_monitor_workspace"} {
		for _, configured := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/raw configured=%t", kind, configured), func(t *testing.T) {
				before := signalConfig(kind, map[string]any{"name": "signal", "signal_definition_name": "definition", "refresh_interval": "PT5M"})
				planned := signalConfig(kind, map[string]any{"name": "signal", "signal_definition_name": "definition", "refresh_interval": "PT5M", "display_name": "changed"})
				rawSignal := map[string]any{"name": "signal", "signal_definition_name": "definition", "display_name": "changed"}
				if configured {
					rawSignal["refresh_interval"] = "PT1M"
				}
				raw := signalConfig(kind, rawSignal)
				wireName := map[string]string{"azure_resource": "azureResource", "azure_log_analytics": "azureLogAnalytics", "azure_monitor_workspace": "azureMonitorWorkspace"}[kind]
				source := map[string]any{"authenticationSetting": "auth", "signals": []any{map[string]any{"name": "signal", "signalDefinitionName": "definition", "refreshInterval": "PT5M"}}}
				switch kind {
				case "azure_resource":
					source["azureResourceId"] = regressionModelID
					source["signals"].([]any)[0].(map[string]any)["signalKind"] = "AzureResourceMetric"
				case "azure_log_analytics":
					source["logAnalyticsWorkspaceResourceId"] = before["signal_group"].([]any)[0].(map[string]any)[kind].([]any)[0].(map[string]any)["log_analytics_workspace_id"]
					source["signals"].([]any)[0].(map[string]any)["signalKind"] = "LogAnalyticsQuery"
				case "azure_monitor_workspace":
					source["azureMonitorWorkspaceResourceId"] = before["signal_group"].([]any)[0].(map[string]any)[kind].([]any)[0].(map[string]any)["monitor_workspace_id"]
					source["signals"].([]any)[0].(map[string]any)["signalKind"] = "PrometheusMetricsQuery"
				}
				encoded, err := json.Marshal(map[string]any{"signalGroups": map[string]any{wireName: source}})
				if err != nil {
					t.Fatal(err)
				}
				props := captureEntityUpdate(t, before, planned, rawConfigValue(t, raw), string(encoded))
				signal := props["signalGroups"].(map[string]any)[wireName].(map[string]any)["signals"].([]any)[0].(map[string]any)
				if signal["refreshInterval"] != "PT5M" || signal["displayName"] != "changed" {
					t.Fatalf("ignored refresh or sibling update lost: %#v", signal)
				}
			})
		}
	}
}

func TestCloudHealthReadNormalizesIDs(t *testing.T) {
	prefix := "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/Group/providers/"
	for _, tc := range []struct {
		field, canonical, wire string
		resource               sdk.Resource
	}{
		{"application_insights_id", prefix + "Microsoft.Insights/components/Component", `{"authenticationSetting":"auth","discoverRelationships":"Enabled","addRecommendedSignals":"Enabled","specification":{"kind":"ApplicationInsightsTopology","applicationInsightsResourceId":%q}}`, HealthModelDiscoveryRuleResource{}},
		{"signal_group.0.azure_log_analytics.0.log_analytics_workspace_id", prefix + "Microsoft.OperationalInsights/workspaces/Workspace", `{"signalGroups":{"azureLogAnalytics":{"authenticationSetting":"auth","logAnalyticsWorkspaceResourceId":%q}}}`, HealthModelEntityResource{}},
		{"signal_group.0.azure_monitor_workspace.0.monitor_workspace_id", prefix + "Microsoft.Monitor/accounts/Workspace", `{"signalGroups":{"azureMonitorWorkspace":{"authenticationSetting":"auth","azureMonitorWorkspaceResourceId":%q}}}`, HealthModelEntityResource{}},
		{"alert.0.unhealthy.0.action_group_ids.0", prefix + "Microsoft.Insights/actionGroups/Action", `{"alerts":{"unhealthy":{"severity":"Sev0","actionGroupIds":[%q]}}}`, HealthModelEntityResource{}},
		{"alert.0.degraded.0.action_group_ids.0", prefix + "Microsoft.Insights/actionGroups/Action", `{"alerts":{"degraded":{"severity":"Sev1","actionGroupIds":[%q]}}}`, HealthModelEntityResource{}},
	} {
		for _, mode := range []string{"mixed case", "malformed", "absent"} {
			t.Run(tc.field+"/"+mode, func(t *testing.T) {
				value := strings.ReplaceAll(strings.ReplaceAll(tc.canonical, "resourceGroups", "RESOURCEGROUPS"), "Microsoft.", "mICROSOFT.")
				properties := fmt.Sprintf(tc.wire, value)
				if mode == "malformed" {
					properties = fmt.Sprintf(tc.wire, "not an ID")
				}
				if mode == "absent" {
					properties = `{}`
					if _, discovery := tc.resource.(HealthModelDiscoveryRuleResource); discovery {
						properties = `{"authenticationSetting":"auth","discoverRelationships":"Enabled","addRecommendedSignals":"Enabled","specification":{"kind":"ResourceGraphQuery","resourceGraphQuery":"resources"}}`
					}
				}
				client := lifecycleClient(t, func(w http.ResponseWriter, request *http.Request) {
					if request.Method != http.MethodGet {
						t.Errorf("unexpected method %s", request.Method)
					}
					w.Header().Set("Content-Type", "application/json")
					_, _ = io.WriteString(w, `{"properties":`+properties+`}`)
				})
				metadata := sdk.NewResourceMetaData(client, tc.resource)
				id := regressionModelID + "/entities/Child"
				if _, discovery := tc.resource.(HealthModelDiscoveryRuleResource); discovery {
					id = regressionModelID + "/discoveryrules/Child"
				}
				metadata.ResourceData.SetId(id)
				err := tc.resource.Read().Func(lifecycleContext(t), metadata)
				if mode == "malformed" {
					if err == nil || !strings.Contains(err.Error(), "parsing") {
						t.Fatalf("expected contextual parse error, got %v", err)
					}
					if metadata.ResourceData.Get("name") != "" {
						t.Fatal("encoded state before validating IDs")
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				if mode != "absent" && metadata.ResourceData.Get(tc.field) != tc.canonical {
					t.Fatalf("expected normalized ID %s, got %v", tc.canonical, metadata.ResourceData.Get(tc.field))
				}
			})
		}
	}
}

func lifecycleContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func captureEntityUpdate(t *testing.T, before, planned map[string]any, raw cty.Value, properties string) map[string]any {
	t.Helper()
	var payload map[string]any
	client := lifecycleClient(t, func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if request.Method == http.MethodGet {
			_, _ = io.WriteString(w, `{"properties":`+properties+`}`)
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
	resource := HealthModelEntityResource{}
	metadata := lifecycleMetadata(t, lifecycleCase{resource: resource, config: before, id: regressionModelID + "/entities/Child"}, planned, raw, client)
	if err := resource.Update().Func(lifecycleContext(t), metadata); err != nil {
		t.Fatal(err)
	}
	return payload["properties"].(map[string]any)
}

func TestCloudHealthUpdateIgnoredDependencyThreshold(t *testing.T) {
	for _, threshold := range []*float64{pointerToFloat(0), pointerToFloat(4), nil} {
		name := "absent"
		if threshold != nil {
			name = fmt.Sprintf("%g", *threshold)
		}
		for _, sibling := range []bool{false, true} {
			t.Run(fmt.Sprintf("threshold=%s/sibling=%t", name, sibling), func(t *testing.T) {
				dependency := map[string]any{"aggregation_type": "MaxNotHealthy", "unit": "Absolute", "unhealthy_threshold": float64(1), "ignore_unknown_enabled": true}
				wire := map[string]any{"aggregationType": "MaxNotHealthy", "unit": "Absolute", "unhealthyThreshold": float64(1), "ignoreUnknown": true}
				if threshold != nil {
					dependency["degraded_threshold"] = *threshold
					wire["degradedThreshold"] = *threshold
				}
				before := map[string]any{"name": "Child", "health_model_id": regressionModelID, "signal_group": []any{map[string]any{"dependency": []any{dependency}}}}
				planned := copyConfig(before)
				plannedDependency := copyConfig(dependency)
				planned["signal_group"] = []any{map[string]any{"dependency": []any{plannedDependency}}}
				if sibling {
					plannedDependency["ignore_unknown_enabled"] = false
				} else {
					planned["display_name"] = "changed"
				}
				rawDependency := copyConfig(plannedDependency)
				delete(rawDependency, "degraded_threshold")
				raw := copyConfig(planned)
				raw["signal_group"] = []any{map[string]any{"dependency": []any{rawDependency}}}
				encoded, err := json.Marshal(wire)
				if err != nil {
					t.Fatal(err)
				}
				properties := captureEntityUpdate(t, before, planned, rawConfigValue(t, raw), `{"signalGroups":{"dependencies":`+string(encoded)+`},"signalAggregationGroups":[{"name":"unmanaged","members":[]}]}`)
				group := properties["signalGroups"].(map[string]any)["dependencies"].(map[string]any)
				actual, present := group["degradedThreshold"]
				if threshold == nil && present || threshold != nil && (!present || actual != *threshold) {
					t.Fatalf("threshold was not preserved: %#v", group)
				}
				if sibling && group["ignoreUnknown"] != false {
					t.Fatalf("sibling change not applied: %#v", group)
				}
				if _, exists := properties["signalAggregationGroups"]; !exists {
					t.Fatal("unmanaged aggregation groups lost")
				}
			})
		}
	}
}

func TestCloudHealthUpdateRemovals(t *testing.T) {
	for _, tc := range lifecycleCases() {
		t.Run(tc.resource.ResourceType(), func(t *testing.T) {
			before := copyConfig(tc.config)
			planned := copyConfig(tc.config)
			var payload map[string]any
			client := lifecycleClient(t, func(w http.ResponseWriter, request *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if request.Method == http.MethodGet {
					_, _ = io.WriteString(w, tc.response)
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
			if _, ok := tc.resource.(HealthModelResource); ok {
				before["identity"] = []any{map[string]any{"type": "SystemAssigned"}}
				before["tags"] = map[string]any{"remove": "tag"}
			} else {
				before["display_name"] = "remove"
				switch tc.resource.(type) {
				case HealthModelEntityResource:
					before["icon"] = []any{map[string]any{"name": "icon"}}
					before["signal_group"] = []any{map[string]any{"dependency": []any{map[string]any{"aggregation_type": "WorstOf"}}}}
					before["health_objective_percentage"] = float64(0)
					before["tags"] = map[string]any{"remove": "tag"}
					tc.response = `{"properties":{"displayName":"remove","healthObjective":0,"icon":{"iconName":"icon"},"signalGroups":{"dependencies":{"aggregationType":"WorstOf"}},"tags":{"remove":"tag"},"signalAggregationGroups":[{"name":"keep","members":[]}]}}`
				case HealthModelRelationshipResource, HealthModelSignalDefinitionResource:
					before["tags"] = map[string]any{"remove": "tag"}
				}
			}
			tc.config = before
			metadata := lifecycleMetadata(t, tc, planned, rawConfigValue(t, planned), client)
			if err := tc.resource.Update().Func(lifecycleContext(t), metadata); err != nil {
				t.Fatal(err)
			}
			if _, ok := tc.resource.(HealthModelResource); ok {
				if payload["identity"].(map[string]any)["type"] != "None" {
					t.Fatalf("identity removal not sent: %#v", payload)
				}
				if !reflect.DeepEqual(payload["tags"], map[string]any{}) {
					t.Fatalf("tags did not clear: %#v", payload)
				}
			} else {
				props := payload["properties"].(map[string]any)
				if _, ok := props["displayName"]; ok {
					t.Fatalf("displayName restored: %#v", props)
				}
				switch tc.resource.(type) {
				case HealthModelEntityResource, HealthModelRelationshipResource, HealthModelSignalDefinitionResource:
					if _, isEntity := tc.resource.(HealthModelEntityResource); isEntity {
						for _, field := range []string{"icon", "signalGroups", "healthObjective"} {
							if _, ok := props[field]; ok {
								t.Fatalf("%s restored: %#v", field, props)
							}
						}
						if _, ok := props["signalAggregationGroups"]; !ok {
							t.Fatal("unmanaged field lost")
						}
					}
					if !reflect.DeepEqual(props["tags"], map[string]any{}) {
						t.Fatalf("tags did not clear: %#v", props)
					}
				}
			}
		})
	}
	for _, threshold := range []float64{0, 3} {
		t.Run(fmt.Sprintf("nested degraded threshold %g", threshold), func(t *testing.T) {
			dependency := map[string]any{"aggregation_type": "MaxNotHealthy", "unit": "Absolute", "unhealthy_threshold": float64(1), "degraded_threshold": threshold}
			before := map[string]any{"name": "Child", "health_model_id": regressionModelID, "signal_group": []any{map[string]any{"dependency": []any{dependency}}}}
			planned := copyConfig(before)
			next := copyConfig(dependency)
			delete(next, "degraded_threshold")
			planned["signal_group"] = []any{map[string]any{"dependency": []any{next}}}
			props := captureEntityUpdate(t, before, planned, rawConfigValue(t, planned), fmt.Sprintf(`{"signalGroups":{"dependencies":{"aggregationType":"MaxNotHealthy","unit":"Absolute","unhealthyThreshold":1,"degradedThreshold":%g}}}`, threshold))
			if _, ok := props["signalGroups"].(map[string]any)["dependencies"].(map[string]any)["degradedThreshold"]; ok {
				t.Fatalf("removed threshold restored: %#v", props)
			}
		})
	}
}
