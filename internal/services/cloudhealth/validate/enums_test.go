// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package validate

import (
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/internal/services/cloudhealth/sdk/2026-11-01/authenticationsettings"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/cloudhealth/sdk/2026-11-01/discoveryrules"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/cloudhealth/sdk/2026-11-01/entities"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/cloudhealth/sdk/2026-11-01/signaldefinitions"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/validation"
)

func TestPossibleValuesAreUsableInStringInSlice(t *testing.T) {
	enums := map[string][]string{
		"AlertSeverity":               entities.PossibleValuesForAlertSeverity(),
		"AuthenticationKind":          authenticationsettings.PossibleValuesForAuthenticationKind(),
		"DependenciesAggregationType": entities.PossibleValuesForAggregationType(),
		"DependenciesAggregationUnit": entities.PossibleValuesForAggregationUnit(),
		"DiscoveryRuleKind":           discoveryrules.PossibleValuesForDiscoveryRuleKind(),
		"DynamicThresholdSensitivity": entities.PossibleValuesForDynamicThresholdSensitivity(),
		"MetricAggregationType":       entities.PossibleValuesForMetricAggregationType(),
		"RefreshInterval":             entities.PossibleValuesForRefreshInterval(),
		"SignalKind":                  entities.PossibleValuesForSignalKind(),
		"SignalOperator":              entities.PossibleValuesForSignalOperator(),
		"EntityImpact":                entities.PossibleValuesForEntityImpact(),
	}

	expectedCounts := map[string]int{
		"AlertSeverity":               5,
		"AuthenticationKind":          1,
		"DependenciesAggregationType": 4,
		"DependenciesAggregationUnit": 2,
		"DiscoveryRuleKind":           2,
		"DynamicThresholdSensitivity": 3,
		"MetricAggregationType":       6,
		"RefreshInterval":             7,
		"SignalKind":                  4,
		"SignalOperator":              7,
		"EntityImpact":                3,
	}

	for name, values := range enums {
		t.Run(name, func(t *testing.T) {
			if len(values) != expectedCounts[name] {
				t.Fatalf("expected %s to have %d values from the 2026-11-01 swagger, got %d: %v", name, expectedCounts[name], len(values), values)
			}

			seen := make(map[string]struct{}, len(values))
			for _, value := range values {
				if strings.TrimSpace(value) != value || value == "" {
					t.Fatalf("%s contains an empty value: %v", name, values)
				}
				if _, duplicated := seen[value]; duplicated {
					t.Fatalf("%s contains the duplicate value %q: %v", name, value, values)
				}
				seen[value] = struct{}{}
				if _, errors := validation.StringInSlice(values, false)(value, name); len(errors) != 0 {
					t.Fatalf("enum member rejected: %v", errors)
				}
			}
			if _, errors := validation.StringInSlice(values, false)("not-an-enum-value", name); len(errors) == 0 {
				t.Fatal("validator accepts an unknown value")
			}
		})
	}
	for name, definitionValues := range map[string][]string{
		"DynamicThresholdSensitivity": signaldefinitions.PossibleValuesForDynamicThresholdSensitivity(),
		"MetricAggregationType":       signaldefinitions.PossibleValuesForMetricAggregationType(),
		"RefreshInterval":             signaldefinitions.PossibleValuesForRefreshInterval(),
		"SignalKind":                  signaldefinitions.PossibleValuesForSignalKind(),
		"SignalOperator":              signaldefinitions.PossibleValuesForSignalOperator(),
	} {
		t.Run(name+"/shared definition set", func(t *testing.T) {
			if !reflect.DeepEqual(enums[name], definitionValues) {
				t.Fatalf("entity and definition enum sets differ: %v, %v", enums[name], definitionValues)
			}
		})
	}
}
