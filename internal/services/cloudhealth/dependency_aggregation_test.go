// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cloudhealth

import (
	"strings"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
)

// TestValidateDependencyAggregation covers both directions of the `dependency` contract the service
// enforces: `WorstOf` refuses thresholds and units, and the counting aggregations require them.
// The guard reads the raw configuration rather than the decoded model because the threshold fields
// are `TypeFloat`, where an explicit `0` - a value the service accepts and stores - decodes
// identically to an absent one.
func TestValidateDependencyAggregation(t *testing.T) {
	testData := []struct {
		name          string
		input         map[string]cty.Value
		expectedError string
	}{
		{
			name:  "worst of without thresholds is accepted",
			input: map[string]cty.Value{"aggregation_type": cty.StringVal("WorstOf")},
		},
		{
			name: "worst of with a unit is rejected",
			input: map[string]cty.Value{
				"aggregation_type": cty.StringVal("WorstOf"),
				"unit":             cty.StringVal("Absolute"),
			},
			expectedError: "`unit` cannot be set",
		},
		{
			name: "worst of with an unhealthy threshold is rejected",
			input: map[string]cty.Value{
				"aggregation_type":    cty.StringVal("WorstOf"),
				"unhealthy_threshold": cty.NumberIntVal(1),
			},
			expectedError: "`unhealthy_threshold` cannot be set",
		},
		{
			name: "worst of with a degraded threshold is rejected",
			input: map[string]cty.Value{
				"aggregation_type":   cty.StringVal("WorstOf"),
				"degraded_threshold": cty.NumberIntVal(2),
			},
			expectedError: "`degraded_threshold` cannot be set",
		},
		{
			// The zero that a decoded model could not tell apart from an absent value.
			name: "worst of with a zero threshold is still rejected",
			input: map[string]cty.Value{
				"aggregation_type":    cty.StringVal("WorstOf"),
				"unhealthy_threshold": cty.NumberIntVal(0),
			},
			expectedError: "`unhealthy_threshold` cannot be set",
		},
		{
			name: "min healthy with both required fields is accepted",
			input: map[string]cty.Value{
				"aggregation_type":    cty.StringVal("MinHealthy"),
				"unit":                cty.StringVal("Absolute"),
				"unhealthy_threshold": cty.NumberIntVal(1),
			},
		},
		{
			// The service stores a threshold of 0, so the guard must not treat it as absent.
			name: "min healthy with a zero threshold is accepted",
			input: map[string]cty.Value{
				"aggregation_type":    cty.StringVal("MinHealthy"),
				"unit":                cty.StringVal("Absolute"),
				"unhealthy_threshold": cty.NumberIntVal(0),
			},
		},
		{
			name: "min healthy without a unit is rejected",
			input: map[string]cty.Value{
				"aggregation_type":    cty.StringVal("MinHealthy"),
				"unhealthy_threshold": cty.NumberIntVal(1),
			},
			expectedError: "`unit` must be set",
		},
		{
			name: "min healthy without an unhealthy threshold is rejected",
			input: map[string]cty.Value{
				"aggregation_type": cty.StringVal("MinHealthy"),
				"unit":             cty.StringVal("Absolute"),
			},
			expectedError: "`unhealthy_threshold` must be set",
		},
		{
			name: "max not healthy without either required field is rejected",
			input: map[string]cty.Value{
				"aggregation_type": cty.StringVal("MaxNotHealthy"),
			},
			expectedError: "must be set",
		},
		{
			name: "max not healthy with both required fields is accepted",
			input: map[string]cty.Value{
				"aggregation_type":    cty.StringVal("MaxNotHealthy"),
				"unit":                cty.StringVal("Percentage"),
				"unhealthy_threshold": cty.NumberIntVal(90),
			},
		},
		{
			// `degraded_threshold` is optional in every mode, so its absence is never an error.
			name: "min healthy without a degraded threshold is accepted",
			input: map[string]cty.Value{
				"aggregation_type":    cty.StringVal("MinHealthy"),
				"unit":                cty.StringVal("Absolute"),
				"unhealthy_threshold": cty.NumberIntVal(1),
				"degraded_threshold":  cty.NullVal(cty.Number),
			},
		},
		{
			// Derived from another resource: nothing to validate until apply.
			name: "an unknown aggregation type is deferred",
			input: map[string]cty.Value{
				"aggregation_type":    cty.UnknownVal(cty.String),
				"unhealthy_threshold": cty.NumberIntVal(1),
			},
		},
	}

	for _, v := range testData {
		t.Run(v.name, func(t *testing.T) {
			err := validateDependencyAggregation(cty.ObjectVal(withNullAttributes(v.input)))

			switch {
			case v.expectedError == "" && err != nil:
				t.Fatalf("expected no error but got: %+v", err)
			case v.expectedError != "" && err == nil:
				t.Fatalf("expected an error containing %q but got none", v.expectedError)
			case v.expectedError != "" && !strings.Contains(err.Error(), v.expectedError):
				t.Fatalf("expected an error containing %q but got: %+v", v.expectedError, err)
			}
		})
	}
}

// withNullAttributes fills in the attributes the case did not set, so every object carries the full
// block shape the schema produces. Terraform supplies an explicit null for an unset attribute
// rather than omitting it, and the guard relies on that.
func withNullAttributes(input map[string]cty.Value) map[string]cty.Value {
	output := map[string]cty.Value{
		"aggregation_type":    cty.NullVal(cty.String),
		"unit":                cty.NullVal(cty.String),
		"unhealthy_threshold": cty.NullVal(cty.Number),
		"degraded_threshold":  cty.NullVal(cty.Number),
	}

	for key, value := range input {
		output[key] = value
	}

	return output
}

// TestExpandHealthModelEntityDependencies asserts the payload half of the same contract: a threshold
// of `0` must survive expansion for the counting aggregations, because the service accepts and
// stores it, while `WorstOf` must send no threshold at all.
func TestExpandHealthModelEntityDependencies(t *testing.T) {
	testData := []struct {
		name                        string
		input                       DependencySignalGroupModel
		degradedThresholdConfigured bool
		expectedUnhealthyThreshold  *float64
		expectedDegradedThreshold   *float64
	}{
		{
			name:                       "min healthy sends a zero unhealthy threshold",
			input:                      DependencySignalGroupModel{AggregationType: "MinHealthy", Unit: "Absolute", UnhealthyThreshold: 0},
			expectedUnhealthyThreshold: pointerToFloat(0),
		},
		{
			name:                       "min healthy sends a non zero unhealthy threshold",
			input:                      DependencySignalGroupModel{AggregationType: "MinHealthy", Unit: "Absolute", UnhealthyThreshold: 3},
			expectedUnhealthyThreshold: pointerToFloat(3),
		},
		{
			name:                        "max not healthy sends both thresholds",
			input:                       DependencySignalGroupModel{AggregationType: "MaxNotHealthy", Unit: "Percentage", UnhealthyThreshold: 90, DegradedThreshold: 50},
			degradedThresholdConfigured: true,
			expectedUnhealthyThreshold:  pointerToFloat(90),
			expectedDegradedThreshold:   pointerToFloat(50),
		},
		{
			name:                        "max not healthy sends a zero degraded threshold",
			input:                       DependencySignalGroupModel{AggregationType: "MaxNotHealthy", Unit: "Absolute", UnhealthyThreshold: 1, DegradedThreshold: 0},
			degradedThresholdConfigured: true,
			expectedUnhealthyThreshold:  pointerToFloat(1),
			expectedDegradedThreshold:   pointerToFloat(0),
		},
		{
			name:                       "an absent degraded threshold is not sent",
			input:                      DependencySignalGroupModel{AggregationType: "MaxNotHealthy", Unit: "Absolute", UnhealthyThreshold: 1},
			expectedUnhealthyThreshold: pointerToFloat(1),
		},
		{
			name:  "worst of sends no thresholds",
			input: DependencySignalGroupModel{AggregationType: "WorstOf"},
		},
	}

	for _, v := range testData {
		t.Run(v.name, func(t *testing.T) {
			actual := expandHealthModelEntityDependencies([]DependencySignalGroupModel{v.input}, v.degradedThresholdConfigured)
			if actual == nil {
				t.Fatal("expected a dependencies signal group but got nil")
			}

			if string(actual.AggregationType) != v.input.AggregationType {
				t.Fatalf("expected aggregation_type %q but got %q", v.input.AggregationType, actual.AggregationType)
			}

			assertFloatPointer(t, "unhealthy_threshold", v.expectedUnhealthyThreshold, actual.UnhealthyThreshold)
			assertFloatPointer(t, "degraded_threshold", v.expectedDegradedThreshold, actual.DegradedThreshold)
		})
	}
}

// TestDependencyDegradedThresholdConfigured pins the distinction the decoded model destroys. A
// nested `TypeFloat` arrives as `0` whether it was configured as `0` or left out entirely, so the
// expander asks the raw configuration instead. Getting this wrong in either direction is a live
// bug: a false positive sends `degradedThreshold: 0` on every dependency block, which the service
// rejects under `WorstOf`, and a false negative silently drops a configured `0`.
func TestDependencyDegradedThresholdConfigured(t *testing.T) {
	dependencyConfig := func(degradedThreshold cty.Value) cty.Value {
		return cty.ObjectVal(map[string]cty.Value{
			"signal_group": cty.ListVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"dependency": cty.ListVal([]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"aggregation_type":    cty.StringVal("MaxNotHealthy"),
							"unit":                cty.StringVal("Absolute"),
							"unhealthy_threshold": cty.NumberIntVal(1),
							"degraded_threshold":  degradedThreshold,
						}),
					}),
				}),
			}),
		})
	}

	testData := []struct {
		name     string
		input    cty.Value
		expected bool
	}{
		{
			name:     "an explicit zero is configured",
			input:    dependencyConfig(cty.NumberIntVal(0)),
			expected: true,
		},
		{
			name:     "an explicit non zero is configured",
			input:    dependencyConfig(cty.NumberIntVal(5)),
			expected: true,
		},
		{
			name:     "a null is not configured",
			input:    dependencyConfig(cty.NullVal(cty.Number)),
			expected: false,
		},
		{
			name:     "an unknown is configured, matching the guards' semantics",
			input:    dependencyConfig(cty.UnknownVal(cty.Number)),
			expected: true,
		},
		{
			name:     "a configuration with no dependency block is not configured",
			input:    cty.ObjectVal(map[string]cty.Value{"name": cty.StringVal("example")}),
			expected: false,
		},
		{
			name:     "a null configuration is not configured",
			input:    cty.NullVal(cty.DynamicPseudoType),
			expected: false,
		},
	}

	for _, v := range testData {
		t.Run(v.name, func(t *testing.T) {
			if actual := dependencyDegradedThresholdConfigured(v.input); actual != v.expected {
				t.Fatalf("expected %t but got %t", v.expected, actual)
			}
		})
	}
}

// TestDependencyBlocksAreSingletons pins the schema property that
// `dependencyDegradedThresholdConfigured` depends on. It answers "is `degraded_threshold` set" once
// for the whole resource, which is indistinguishable from "is it set on this block" only while at
// most one `dependency` block can exist. Raising either `MaxItems` would make one configured
// threshold silently enable the field on every sibling block, so the coupling is asserted rather
// than left to a comment.
func TestDependencyBlocksAreSingletons(t *testing.T) {
	resourceSchema := HealthModelEntityResource{}.Arguments()

	signalGroup, ok := resourceSchema["signal_group"]
	if !ok {
		t.Fatal("expected a `signal_group` argument")
	}

	if signalGroup.MaxItems != 1 {
		t.Fatalf("expected `signal_group` to allow at most one block but MaxItems is %d - `dependencyDegradedThresholdConfigured` returns a single answer for the whole resource and would now apply one block's `degraded_threshold` to all of them", signalGroup.MaxItems)
	}

	signalGroupResource, ok := signalGroup.Elem.(*pluginsdk.Resource)
	if !ok {
		t.Fatalf("expected `signal_group` to nest a resource but got %T", signalGroup.Elem)
	}

	dependency, ok := signalGroupResource.Schema["dependency"]
	if !ok {
		t.Fatal("expected a `signal_group.dependency` argument")
	}

	if dependency.MaxItems != 1 {
		t.Fatalf("expected `signal_group.dependency` to allow at most one block but MaxItems is %d - `dependencyDegradedThresholdConfigured` returns a single answer for the whole resource and would now apply one block's `degraded_threshold` to all of them", dependency.MaxItems)
	}
}
