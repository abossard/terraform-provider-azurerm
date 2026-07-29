// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package validate

import (
	"testing"
)

// TestPossibleValuesAreUsableInStringInSlice guards the transcribed swagger enums against the failure
// modes that a hand-maintained copy actually has: an empty set, a duplicated entry from a copy/paste,
// or a blank/whitespace value. Each of these silently produces a validator that accepts or rejects the
// wrong things.
func TestPossibleValuesAreUsableInStringInSlice(t *testing.T) {
	enums := map[string][]string{
		"AlertSeverity":               PossibleValuesForAlertSeverity(),
		"AuthenticationKind":          PossibleValuesForAuthenticationKind(),
		"DependenciesAggregationType": PossibleValuesForDependenciesAggregationType(),
		"DependenciesAggregationUnit": PossibleValuesForDependenciesAggregationUnit(),
		"DiscoveryRuleKind":           PossibleValuesForDiscoveryRuleKind(),
		"DynamicThresholdSensitivity": PossibleValuesForDynamicThresholdSensitivity(),
		"LookBackWindow":              PossibleValuesForLookBackWindow(),
		"MetricAggregationType":       PossibleValuesForMetricAggregationType(),
		"RefreshInterval":             PossibleValuesForRefreshInterval(),
		"SignalKind":                  PossibleValuesForSignalKind(),
		"SignalOperator":              PossibleValuesForSignalOperator(),
	}

	expectedCounts := map[string]int{
		"AlertSeverity":               5,
		"AuthenticationKind":          1,
		"DependenciesAggregationType": 3,
		"DependenciesAggregationUnit": 2,
		"DiscoveryRuleKind":           2,
		"DynamicThresholdSensitivity": 3,
		"LookBackWindow":              4,
		"MetricAggregationType":       6,
		"RefreshInterval":             7,
		"SignalKind":                  4,
		"SignalOperator":              7,
	}

	for name, values := range enums {
		t.Run(name, func(t *testing.T) {
			if len(values) != expectedCounts[name] {
				t.Fatalf("expected %s to have %d values from the 2026-05-01-preview swagger, got %d: %v", name, expectedCounts[name], len(values), values)
			}

			seen := make(map[string]struct{}, len(values))
			for _, value := range values {
				if value == "" {
					t.Fatalf("%s contains an empty value: %v", name, values)
				}
				if _, duplicated := seen[value]; duplicated {
					t.Fatalf("%s contains the duplicate value %q: %v", name, value, values)
				}
				seen[value] = struct{}{}
			}
		})
	}
}
