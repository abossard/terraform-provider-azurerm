// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package validate

// The enum value sets below are transcribed from the `2026-11-01` Microsoft.CloudHealth
// swagger at `Azure/azure-rest-api-specs`. The function names mirror the SDK convention so each
// call site can be repointed when CloudHealth is available in the vendored `hashicorp/go-azure-sdk`.

// PossibleValuesForAuthenticationKind returns the values of `#/definitions/AuthenticationKind`.
func PossibleValuesForAuthenticationKind() []string {
	return []string{
		"ManagedIdentity",
	}
}

// PossibleValuesForEntityImpact returns the values of `#/definitions/EntityProperties/properties/impact`.
func PossibleValuesForEntityImpact() []string {
	return []string{
		"Limited",
		"Standard",
		"Suppressed",
	}
}

// PossibleValuesForDependenciesAggregationType returns the values of
// `#/definitions/DependenciesSignalGroupV2/properties/aggregationType`.
func PossibleValuesForDependenciesAggregationType() []string {
	return []string{
		"BestOf",
		"MaxNotHealthy",
		"MinHealthy",
		"WorstOf",
	}
}

// PossibleValuesForDependenciesAggregationUnit returns the values of `#/definitions/AggregationUnit`.
func PossibleValuesForDependenciesAggregationUnit() []string {
	return []string{
		"Absolute",
		"Percentage",
	}
}

// PossibleValuesForMetricAggregationType returns the values of `#/definitions/MetricAggregationType`.
func PossibleValuesForMetricAggregationType() []string {
	return []string{
		"Average",
		"Count",
		"Maximum",
		"Minimum",
		"None",
		"Total",
	}
}

// PossibleValuesForRefreshInterval returns the values of the `refreshInterval` enum shared by
// `#/definitions/SignalDefinitionProperties`, `AzureResourceSignal`, `LogAnalyticsSignal` and
// `PrometheusMetricsSignal`.
func PossibleValuesForRefreshInterval() []string {
	return []string{
		"PT1H",
		"PT1M",
		"PT2H",
		"PT5M",
		"PT10M",
		"PT15M",
		"PT30M",
	}
}

// PossibleValuesForSignalKind returns the values of `#/definitions/SignalKind`.
func PossibleValuesForSignalKind() []string {
	return []string{
		"AzureResourceMetric",
		"External",
		"LogAnalyticsQuery",
		"PrometheusMetricsQuery",
	}
}

// PossibleValuesForSignalOperator returns the values of `#/definitions/SignalOperator`.
func PossibleValuesForSignalOperator() []string {
	return []string{
		"Dynamic",
		"Equal",
		"GreaterThan",
		"GreaterThanOrEqual",
		"LessThan",
		"LessThanOrEqual",
		"NotEqual",
	}
}

// PossibleValuesForDynamicThresholdSensitivity returns the values of `#/definitions/DynamicThresholdSensitivity`.
func PossibleValuesForDynamicThresholdSensitivity() []string {
	return []string{
		"High",
		"Low",
		"Medium",
	}
}

// PossibleValuesForAlertSeverity returns the values of `#/definitions/AlertSeverity`.
func PossibleValuesForAlertSeverity() []string {
	return []string{
		"Sev0",
		"Sev1",
		"Sev2",
		"Sev3",
		"Sev4",
	}
}

// PossibleValuesForDiscoveryRuleKind returns the values of `#/definitions/DiscoveryRuleKind`.
func PossibleValuesForDiscoveryRuleKind() []string {
	return []string{
		"ApplicationInsightsTopology",
		"ResourceGraphQuery",
	}
}
