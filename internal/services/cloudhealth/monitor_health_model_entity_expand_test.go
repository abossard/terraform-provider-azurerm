// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cloudhealth

import (
	"testing"
)

// TestExpandHealthModelEntityProperties pins the zero-versus-absent distinction for
// `health_objective_percentage`. The service stores an explicit `0` and reads an omitted value back
// as absent, so a conditional expander keyed on the value would silently discard a documented,
// schema-valid configuration.
func TestExpandHealthModelEntityProperties(t *testing.T) {
	testData := []struct {
		name                    string
		input                   HealthModelEntityResourceModel
		expectedHealthObjective *float64
		expectedDisplayName     *string
		expectedImpact          bool
	}{
		{
			name:                    "a zero health objective is forwarded",
			input:                   HealthModelEntityResourceModel{HealthObjectivePercentage: pointerToFloat(0)},
			expectedHealthObjective: pointerToFloat(0),
		},
		{
			name:                    "a non zero health objective is forwarded",
			input:                   HealthModelEntityResourceModel{HealthObjectivePercentage: pointerToFloat(99.9)},
			expectedHealthObjective: pointerToFloat(99.9),
		},
		{
			name:                    "an absent health objective is not sent",
			input:                   HealthModelEntityResourceModel{},
			expectedHealthObjective: nil,
		},
		{
			name:                    "the other optional properties are unaffected",
			input:                   HealthModelEntityResourceModel{HealthObjectivePercentage: pointerToFloat(0), DisplayName: "Payments", Impact: "Standard"},
			expectedHealthObjective: pointerToFloat(0),
			expectedDisplayName:     pointerToString("Payments"),
			expectedImpact:          true,
		},
	}

	for _, v := range testData {
		t.Run(v.name, func(t *testing.T) {
			actual := expandHealthModelEntityProperties(v.input, false)
			if actual == nil {
				t.Fatal("expected entity properties but got nil")
			}

			assertFloatPointer(t, "health_objective_percentage", v.expectedHealthObjective, actual.HealthObjective)

			switch {
			case v.expectedDisplayName == nil && actual.DisplayName != nil:
				t.Fatalf("expected display_name to be absent but got %q", *actual.DisplayName)
			case v.expectedDisplayName != nil && actual.DisplayName == nil:
				t.Fatalf("expected display_name %q but it was absent", *v.expectedDisplayName)
			case v.expectedDisplayName != nil && *actual.DisplayName != *v.expectedDisplayName:
				t.Fatalf("expected display_name %q but got %q", *v.expectedDisplayName, *actual.DisplayName)
			}

			if got := actual.Impact != nil; got != v.expectedImpact {
				t.Fatalf("expected impact present to be %t but got %t", v.expectedImpact, got)
			}
		})
	}
}
