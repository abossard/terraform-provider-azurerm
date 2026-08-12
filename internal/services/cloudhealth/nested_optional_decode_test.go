// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cloudhealth

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/internal/sdk"
)

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
