// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cloudhealth

import (
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/internal/services/cloudhealth/sdk/2026-11-01/entities"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
)

// TestExpandThresholdRule covers the apply-time half of the `evaluation_rule` contract. `CustomizeDiff`
// rejects a mismatched `operator` at plan time, but only when `operator` is known then - a value derived
// from another resource is unknown until apply. This asserts that the expander forwards whatever was
// configured in that case, so the service can reject it by name, rather than dropping it and leaving a
// permanent diff that no subsequent apply can settle.
func TestExpandThresholdRule(t *testing.T) {
	testData := []struct {
		name               string
		input              ThresholdRuleModel
		expectedThreshold  *float64
		expectedSensitvity *entities.DynamicThresholdSensitivity
	}{
		{
			name:              "static operator sends threshold",
			input:             ThresholdRuleModel{Operator: "LessThan", Threshold: 99},
			expectedThreshold: pointerToFloat(99),
		},
		{
			name:              "static operator sends an explicit zero threshold",
			input:             ThresholdRuleModel{Operator: "LessThan", Threshold: 0},
			expectedThreshold: pointerToFloat(0),
		},
		{
			name:               "dynamic operator sends sensitivity",
			input:              ThresholdRuleModel{Operator: "Dynamic", Sensitivity: "Medium"},
			expectedSensitvity: pointerToSensitivity(entities.DynamicThresholdSensitivityMedium),
		},
		{
			name:  "dynamic operator omits an absent threshold",
			input: ThresholdRuleModel{Operator: "Dynamic", Sensitivity: "Medium"},
			// A decoded zero is indistinguishable from an absent value, and `CustomizeDiff` already
			// rejects an explicit `threshold = 0` here, so nothing is sent.
			expectedThreshold:  nil,
			expectedSensitvity: pointerToSensitivity(entities.DynamicThresholdSensitivityMedium),
		},
		{
			name: "dynamic operator forwards a configured threshold for the service to reject",
			// Only reachable when `operator` is unknown at plan time. Forwarding produces
			// `DynamicThresholdNotAllowed`; dropping would produce a permanent diff.
			input:             ThresholdRuleModel{Operator: "Dynamic", Threshold: 99},
			expectedThreshold: pointerToFloat(99),
		},
		{
			name: "static operator forwards a configured sensitivity for the service to reject",
			input: ThresholdRuleModel{
				Operator: "LessThan", Threshold: 99, Sensitivity: "Medium",
			},
			expectedThreshold:  pointerToFloat(99),
			expectedSensitvity: pointerToSensitivity(entities.DynamicThresholdSensitivityMedium),
		},
	}

	for _, v := range testData {
		t.Run(v.name, func(t *testing.T) {
			actual := expandThresholdRule(v.input)

			if string(actual.Operator) != v.input.Operator {
				t.Fatalf("expected operator %q but got %q", v.input.Operator, actual.Operator)
			}

			assertFloatPointer(t, "threshold", v.expectedThreshold, actual.Threshold)

			var actualSensitivity *string
			if actual.Sensitivity != nil {
				actualSensitivity = pointerToString(string(*actual.Sensitivity))
			}
			var expectedSensitivity *string
			if v.expectedSensitvity != nil {
				expectedSensitivity = pointerToString(string(*v.expectedSensitvity))
			}
			assertStringPointer(t, "sensitivity", expectedSensitivity, actualSensitivity)
		})
	}
}

func TestThresholdRuleSchemaStableAPI(t *testing.T) {
	testData := []struct {
		name     string
		required bool
	}{
		{name: "unhealthy", required: true},
		{name: "degraded", required: false},
	}

	for _, v := range testData {
		t.Run(v.name, func(t *testing.T) {
			schema := thresholdRuleSchema(v.required)
			resource, ok := schema.Elem.(*pluginsdk.Resource)
			if !ok {
				t.Fatalf("expected a threshold resource but got %T", schema.Elem)
			}

			if _, exists := resource.Schema["look_back_window"]; exists {
				t.Fatal("look_back_window is not supported by the 2026-11-01 stable API")
			}
		})
	}
}

func assertFloatPointer(t *testing.T, field string, expected, actual *float64) {
	t.Helper()

	switch {
	case expected == nil && actual != nil:
		t.Fatalf("expected %s to be omitted but got %f", field, *actual)
	case expected != nil && actual == nil:
		t.Fatalf("expected %s to be %f but it was omitted", field, *expected)
	case expected != nil && *expected != *actual:
		t.Fatalf("expected %s to be %f but got %f", field, *expected, *actual)
	}
}

func assertStringPointer(t *testing.T, field string, expected, actual *string) {
	t.Helper()

	switch {
	case expected == nil && actual != nil:
		t.Fatalf("expected %s to be omitted but got %q", field, *actual)
	case expected != nil && actual == nil:
		t.Fatalf("expected %s to be %q but it was omitted", field, *expected)
	case expected != nil && *expected != *actual:
		t.Fatalf("expected %s to be %q but got %q", field, *expected, *actual)
	}
}

func pointerToFloat(input float64) *float64 {
	return &input
}

func pointerToString(input string) *string {
	return &input
}

func pointerToSensitivity(input entities.DynamicThresholdSensitivity) *entities.DynamicThresholdSensitivity {
	return &input
}
