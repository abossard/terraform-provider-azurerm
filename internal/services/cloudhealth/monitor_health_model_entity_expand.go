// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cloudhealth

import (
	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/cloudhealth/sdk/2026-05-01-preview/entities"
)

func expandHealthModelEntityProperties(config HealthModelEntityResourceModel, degradedThresholdConfigured bool) *entities.EntityProperties {
	props := entities.EntityProperties{
		CanvasPosition: expandHealthModelEntityCanvasPosition(config.CanvasPosition),
		Icon:           expandHealthModelEntityIcon(config.Icon),
		SignalGroups:   expandHealthModelEntitySignalGroups(config.SignalGroup, degradedThresholdConfigured),
		Alerts:         expandHealthModelEntityAlerts(config.Alert),
		Tags:           pointer.To(config.Tags),
	}

	if config.DisplayName != "" {
		props.DisplayName = pointer.To(config.DisplayName)
	}

	// `health_objective_percentage` is a pointer in the model so that a configured `0` is
	// distinguishable from an absent one. The service stores an explicit `0` and reads an omitted
	// value back as absent, so the two are not interchangeable and the value must be forwarded
	// exactly as configured.
	props.HealthObjective = config.HealthObjectivePercentage

	if config.Impact != "" {
		props.Impact = pointer.To(entities.EntityImpact(config.Impact))
	}

	return &props
}

func expandHealthModelEntityCanvasPosition(input []CanvasPositionModel) *entities.EntityCoordinates {
	if len(input) == 0 {
		return nil
	}

	return &entities.EntityCoordinates{
		X: input[0].X,
		Y: input[0].Y,
	}
}

func flattenHealthModelEntityCanvasPosition(input *entities.EntityCoordinates) []CanvasPositionModel {
	if input == nil {
		return []CanvasPositionModel{}
	}

	return []CanvasPositionModel{{
		X: input.X,
		Y: input.Y,
	}}
}

func expandHealthModelEntityIcon(input []IconModel) *entities.IconDefinition {
	if len(input) == 0 {
		return nil
	}

	output := entities.IconDefinition{
		IconName: input[0].Name,
	}

	if input[0].CustomDataBase64 != "" {
		output.CustomData = pointer.To(input[0].CustomDataBase64)
	}

	return &output
}

func flattenHealthModelEntityIcon(input *entities.IconDefinition) []IconModel {
	if input == nil {
		return []IconModel{}
	}

	return []IconModel{{
		Name:             input.IconName,
		CustomDataBase64: pointer.From(input.CustomData),
	}}
}

func expandHealthModelEntitySignalGroups(input []SignalGroupModel, degradedThresholdConfigured bool) *entities.SignalGroups {
	if len(input) == 0 {
		return nil
	}

	config := input[0]

	return &entities.SignalGroups{
		AzureResource:         expandHealthModelEntityAzureResourceSignals(config.AzureResource),
		AzureLogAnalytics:     expandHealthModelEntityLogAnalyticsSignals(config.AzureLogAnalytics),
		AzureMonitorWorkspace: expandHealthModelEntityAzureMonitorWorkspaceSignals(config.AzureMonitorWorkspace),
		Dependencies:          expandHealthModelEntityDependencies(config.Dependency, degradedThresholdConfigured),
	}
}

func flattenHealthModelEntitySignalGroups(input *entities.SignalGroups) []SignalGroupModel {
	if input == nil {
		return []SignalGroupModel{}
	}

	output := SignalGroupModel{
		AzureResource:         flattenHealthModelEntityAzureResourceSignals(input.AzureResource),
		AzureLogAnalytics:     flattenHealthModelEntityLogAnalyticsSignals(input.AzureLogAnalytics),
		AzureMonitorWorkspace: flattenHealthModelEntityAzureMonitorWorkspaceSignals(input.AzureMonitorWorkspace),
		Dependency:            flattenHealthModelEntityDependencies(input.Dependencies),
	}

	// The `external` signal group is populated by the service and has no Terraform representation,
	// so a response that carries nothing else must still flatten to an empty block rather than one
	// holding four empty lists.
	if len(output.AzureResource) == 0 && len(output.AzureLogAnalytics) == 0 && len(output.AzureMonitorWorkspace) == 0 && len(output.Dependency) == 0 {
		return []SignalGroupModel{}
	}

	return []SignalGroupModel{output}
}

func expandHealthModelEntityAzureResourceSignals(input []AzureResourceSignalGroupModel) *entities.AzureResourceSignals {
	if len(input) == 0 {
		return nil
	}

	config := input[0]

	resourceHealth := entities.ResourceHealthAvailabilityStateSignalBehaviorDisabled
	if config.ResourceHealthEnabled {
		resourceHealth = entities.ResourceHealthAvailabilityStateSignalBehaviorEnabled
	}

	output := entities.AzureResourceSignals{
		AuthenticationSetting: config.AuthenticationSettingName,
		AzureResourceId:       config.TargetResourceId,
		ResourceHealth: &entities.AzureResourceHealthSignal{
			Enabled: pointer.To(resourceHealth),
		},
	}

	signals := make([]entities.AzureResourceSignal, 0, len(config.Signal))
	for _, signal := range config.Signal {
		item := entities.AzureResourceSignal{
			Name:            signal.Name,
			SignalKind:      entities.SignalKindAzureResourceMetric,
			MetricNamespace: pointer.To(signal.MetricNamespace),
			MetricName:      pointer.To(signal.MetricName),
			AggregationType: pointer.To(entities.MetricAggregationType(signal.AggregationType)),
			TimeGrain:       pointer.To(signal.TimeGrain),
			EvaluationRules: expandEvaluationRule(signal.EvaluationRule),
		}

		if signal.DisplayName != "" {
			item.DisplayName = pointer.To(signal.DisplayName)
		}
		if signal.DimensionFilter != "" {
			item.DimensionFilter = pointer.To(signal.DimensionFilter)
		}
		if signal.RefreshInterval != "" {
			item.RefreshInterval = pointer.To(entities.RefreshInterval(signal.RefreshInterval))
		}
		if signal.DataUnit != "" {
			item.DataUnit = pointer.To(signal.DataUnit)
		}
		if signal.SignalDefinitionName != "" {
			item.SignalDefinitionName = pointer.To(signal.SignalDefinitionName)
		}

		signals = append(signals, item)
	}
	output.Signals = &signals

	return &output
}

func flattenHealthModelEntityAzureResourceSignals(input *entities.AzureResourceSignals) []AzureResourceSignalGroupModel {
	if input == nil {
		return []AzureResourceSignalGroupModel{}
	}

	output := AzureResourceSignalGroupModel{
		AuthenticationSettingName: input.AuthenticationSetting,
		TargetResourceId:          input.AzureResourceId,
		// The service omits `resourceHealth` entirely when the behaviour is its default, so an
		// absent block reads back as enabled rather than as false.
		ResourceHealthEnabled: true,
		Signal:                make([]AzureResourceSignalModel, 0),
	}

	if input.ResourceHealth != nil && input.ResourceHealth.Enabled != nil {
		output.ResourceHealthEnabled = *input.ResourceHealth.Enabled == entities.ResourceHealthAvailabilityStateSignalBehaviorEnabled
	}

	for _, signal := range pointer.From(input.Signals) {
		item := AzureResourceSignalModel{
			Name:                 signal.Name,
			MetricNamespace:      pointer.From(signal.MetricNamespace),
			MetricName:           pointer.From(signal.MetricName),
			TimeGrain:            pointer.From(signal.TimeGrain),
			DisplayName:          pointer.From(signal.DisplayName),
			DimensionFilter:      pointer.From(signal.DimensionFilter),
			DataUnit:             pointer.From(signal.DataUnit),
			SignalDefinitionName: pointer.From(signal.SignalDefinitionName),
			EvaluationRule:       flattenEvaluationRule(signal.EvaluationRules),
		}

		if signal.AggregationType != nil {
			item.AggregationType = string(*signal.AggregationType)
		}
		if signal.RefreshInterval != nil {
			item.RefreshInterval = string(*signal.RefreshInterval)
		}

		output.Signal = append(output.Signal, item)
	}

	return []AzureResourceSignalGroupModel{output}
}

func expandHealthModelEntityLogAnalyticsSignals(input []LogAnalyticsSignalGroupModel) *entities.LogAnalyticsSignals {
	if len(input) == 0 {
		return nil
	}

	config := input[0]

	output := entities.LogAnalyticsSignals{
		AuthenticationSetting:           config.AuthenticationSettingName,
		LogAnalyticsWorkspaceResourceId: config.LogAnalyticsWorkspaceId,
	}

	signals := make([]entities.LogAnalyticsSignal, 0, len(config.Signal))
	for _, signal := range config.Signal {
		item := entities.LogAnalyticsSignal{
			Name:            signal.Name,
			SignalKind:      entities.SignalKindLogAnalyticsQuery,
			QueryText:       pointer.To(signal.QueryText),
			EvaluationRules: expandEvaluationRule(signal.EvaluationRule),
		}

		if signal.TimeGrain != "" {
			item.TimeGrain = pointer.To(signal.TimeGrain)
		}
		if signal.ValueColumnName != "" {
			item.ValueColumnName = pointer.To(signal.ValueColumnName)
		}
		if signal.DisplayName != "" {
			item.DisplayName = pointer.To(signal.DisplayName)
		}
		if signal.RefreshInterval != "" {
			item.RefreshInterval = pointer.To(entities.RefreshInterval(signal.RefreshInterval))
		}
		if signal.DataUnit != "" {
			item.DataUnit = pointer.To(signal.DataUnit)
		}
		if signal.SignalDefinitionName != "" {
			item.SignalDefinitionName = pointer.To(signal.SignalDefinitionName)
		}

		signals = append(signals, item)
	}
	output.Signals = &signals

	return &output
}

func flattenHealthModelEntityLogAnalyticsSignals(input *entities.LogAnalyticsSignals) []LogAnalyticsSignalGroupModel {
	if input == nil {
		return []LogAnalyticsSignalGroupModel{}
	}

	output := LogAnalyticsSignalGroupModel{
		AuthenticationSettingName: input.AuthenticationSetting,
		LogAnalyticsWorkspaceId:   input.LogAnalyticsWorkspaceResourceId,
		Signal:                    make([]LogAnalyticsSignalModel, 0),
	}

	for _, signal := range pointer.From(input.Signals) {
		item := LogAnalyticsSignalModel{
			Name:                 signal.Name,
			QueryText:            pointer.From(signal.QueryText),
			TimeGrain:            pointer.From(signal.TimeGrain),
			ValueColumnName:      pointer.From(signal.ValueColumnName),
			DisplayName:          pointer.From(signal.DisplayName),
			DataUnit:             pointer.From(signal.DataUnit),
			SignalDefinitionName: pointer.From(signal.SignalDefinitionName),
			EvaluationRule:       flattenEvaluationRule(signal.EvaluationRules),
		}

		if signal.RefreshInterval != nil {
			item.RefreshInterval = string(*signal.RefreshInterval)
		}

		output.Signal = append(output.Signal, item)
	}

	return []LogAnalyticsSignalGroupModel{output}
}

func expandHealthModelEntityAzureMonitorWorkspaceSignals(input []AzureMonitorWorkspaceSignalGroupModel) *entities.AzureMonitorWorkspaceSignals {
	if len(input) == 0 {
		return nil
	}

	config := input[0]

	output := entities.AzureMonitorWorkspaceSignals{
		AuthenticationSetting:           config.AuthenticationSettingName,
		AzureMonitorWorkspaceResourceId: config.MonitorWorkspaceId,
	}

	signals := make([]entities.PrometheusMetricsSignal, 0, len(config.Signal))
	for _, signal := range config.Signal {
		item := entities.PrometheusMetricsSignal{
			Name:            signal.Name,
			SignalKind:      entities.SignalKindPrometheusMetricsQuery,
			QueryText:       pointer.To(signal.QueryText),
			EvaluationRules: expandEvaluationRule(signal.EvaluationRule),
		}

		if signal.TimeGrain != "" {
			item.TimeGrain = pointer.To(signal.TimeGrain)
		}
		if signal.DisplayName != "" {
			item.DisplayName = pointer.To(signal.DisplayName)
		}
		if signal.RefreshInterval != "" {
			item.RefreshInterval = pointer.To(entities.RefreshInterval(signal.RefreshInterval))
		}
		if signal.DataUnit != "" {
			item.DataUnit = pointer.To(signal.DataUnit)
		}
		if signal.SignalDefinitionName != "" {
			item.SignalDefinitionName = pointer.To(signal.SignalDefinitionName)
		}

		signals = append(signals, item)
	}
	output.Signals = &signals

	return &output
}

func flattenHealthModelEntityAzureMonitorWorkspaceSignals(input *entities.AzureMonitorWorkspaceSignals) []AzureMonitorWorkspaceSignalGroupModel {
	if input == nil {
		return []AzureMonitorWorkspaceSignalGroupModel{}
	}

	output := AzureMonitorWorkspaceSignalGroupModel{
		AuthenticationSettingName: input.AuthenticationSetting,
		MonitorWorkspaceId:        input.AzureMonitorWorkspaceResourceId,
		Signal:                    make([]PrometheusMetricsSignalModel, 0),
	}

	for _, signal := range pointer.From(input.Signals) {
		item := PrometheusMetricsSignalModel{
			Name:                 signal.Name,
			QueryText:            pointer.From(signal.QueryText),
			TimeGrain:            pointer.From(signal.TimeGrain),
			DisplayName:          pointer.From(signal.DisplayName),
			DataUnit:             pointer.From(signal.DataUnit),
			SignalDefinitionName: pointer.From(signal.SignalDefinitionName),
			EvaluationRule:       flattenEvaluationRule(signal.EvaluationRules),
		}

		if signal.RefreshInterval != nil {
			item.RefreshInterval = string(*signal.RefreshInterval)
		}

		output.Signal = append(output.Signal, item)
	}

	return []AzureMonitorWorkspaceSignalGroupModel{output}
}

func expandHealthModelEntityDependencies(input []DependencySignalGroupModel, degradedThresholdConfigured bool) *entities.DependenciesSignalGroupV2 {
	if len(input) == 0 {
		return nil
	}

	config := input[0]

	output := entities.DependenciesSignalGroupV2{
		AggregationType: entities.DependenciesAggregationType(config.AggregationType),
		IgnoreUnknown:   pointer.To(config.IgnoreUnknownEnabled),
	}

	if config.Unit != "" {
		output.Unit = pointer.To(entities.DependenciesAggregationUnit(config.Unit))
	}

	// `unhealthy_threshold` is `TypeFloat`, so a decoded `0` is indistinguishable from an absent
	// one - and the service both accepts and stores a threshold of `0`. Keying off the aggregation
	// type instead recovers the distinction: the plan-time guard has already established that the
	// counting aggregations always carry a threshold and that `WorstOf` never does, so the
	// aggregation type says which case this is.
	if config.AggregationType != string(entities.DependenciesAggregationTypeWorstOf) {
		output.UnhealthyThreshold = pointer.To(config.UnhealthyThreshold)
	}

	// `degraded_threshold` is `TypeFloat` inside a nested block, so a decoded `0` cannot be told
	// apart from an absent one and a pointer model field does not help: the decoder sets every
	// field of a nested block it visits, so absence arrives as a pointer to `0`. The service treats
	// the two as different resources - `MaxNotHealthy` accepts and stores a degraded threshold of
	// `0`, while an omitted one reads back as absent and means "no degraded state" - so the raw
	// configuration decides whether to send it.
	if degradedThresholdConfigured {
		output.DegradedThreshold = pointer.To(config.DegradedThreshold)
	}

	return &output
}

func flattenHealthModelEntityDependencies(input *entities.DependenciesSignalGroupV2) []DependencySignalGroupModel {
	if input == nil {
		return []DependencySignalGroupModel{}
	}

	output := DependencySignalGroupModel{
		AggregationType:      string(input.AggregationType),
		UnhealthyThreshold:   pointer.From(input.UnhealthyThreshold),
		DegradedThreshold:    pointer.From(input.DegradedThreshold),
		IgnoreUnknownEnabled: pointer.From(input.IgnoreUnknown),
	}

	if input.Unit != nil {
		output.Unit = string(*input.Unit)
	}

	return []DependencySignalGroupModel{output}
}

func expandHealthModelEntityAlerts(input []EntityAlertModel) *entities.EntityAlerts {
	if len(input) == 0 {
		return nil
	}

	return &entities.EntityAlerts{
		Unhealthy: expandHealthModelEntityAlertConfiguration(input[0].Unhealthy),
		Degraded:  expandHealthModelEntityAlertConfiguration(input[0].Degraded),
	}
}

func flattenHealthModelEntityAlerts(input *entities.EntityAlerts) []EntityAlertModel {
	if input == nil {
		return []EntityAlertModel{}
	}

	return []EntityAlertModel{{
		Unhealthy: flattenHealthModelEntityAlertConfiguration(input.Unhealthy),
		Degraded:  flattenHealthModelEntityAlertConfiguration(input.Degraded),
	}}
}

func expandHealthModelEntityAlertConfiguration(input []AlertConfigurationModel) *entities.AlertConfiguration {
	if len(input) == 0 {
		return nil
	}

	output := entities.AlertConfiguration{
		Severity: entities.AlertSeverity(input[0].Severity),
	}

	if input[0].Description != "" {
		output.Description = pointer.To(input[0].Description)
	}

	if len(input[0].ActionGroupIds) > 0 {
		output.ActionGroupIds = pointer.To(input[0].ActionGroupIds)
	}

	return &output
}

func flattenHealthModelEntityAlertConfiguration(input *entities.AlertConfiguration) []AlertConfigurationModel {
	if input == nil {
		return []AlertConfigurationModel{}
	}

	return []AlertConfigurationModel{{
		Severity:       string(input.Severity),
		Description:    pointer.From(input.Description),
		ActionGroupIds: pointer.From(input.ActionGroupIds),
	}}
}
