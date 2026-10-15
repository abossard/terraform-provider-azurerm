// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cloudhealth

import (
	"fmt"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/go-azure-sdk/resource-manager/insights/2023-01-01/actiongroupsapis"
	"github.com/hashicorp/go-azure-sdk/resource-manager/monitor/2023-04-03/azuremonitorworkspaces"
	"github.com/hashicorp/go-azure-sdk/resource-manager/operationalinsights/2020-08-01/workspaces"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/cloudhealth/sdk/2026-11-01/entities"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/ctyhelpers"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
)

func cloudHealthChanges(data *pluginsdk.ResourceData) func(string) bool {
	return func(path string) bool {
		if data.HasChange(path) {
			return true
		}
		// The flat map erases null versus zero; resolved state and plan retain presence changes.
		state, plan := data.GetRawState(), data.GetRawPlan()
		if state.IsNull() || plan.IsNull() || !state.IsKnown() || !plan.IsKnown() {
			return false
		}
		ctyPath := ctyhelpers.ConstructCtyPath(path)
		before, beforeErr := ctyPath.Apply(state)
		after, afterErr := ctyPath.Apply(plan)
		beforeAbsent := beforeErr != nil || before.IsNull()
		afterAbsent := afterErr != nil || after.IsNull()
		if beforeAbsent || afterAbsent {
			return beforeAbsent != afterAbsent
		}
		return before.IsWhollyKnown() && after.IsWhollyKnown() && !before.RawEquals(after)
	}
}

func changedValue[T any](hasChange func(string) bool, path string, current, planned T) T {
	if hasChange(path) {
		return planned
	}
	return current
}

func healthModelEntityRefreshOverrides(config *HealthModelEntityResourceModel, raw cty.Value) {
	if raw.IsNull() || !raw.IsKnown() {
		return
	}
	configured := func(kind string, i int) bool {
		value, err := ctyhelpers.ConstructCtyPath(fmt.Sprintf("signal_group.0.%s.0.signal.%d.refresh_interval", kind, i)).Apply(raw)
		return err == nil && !value.IsNull()
	}
	for _, group := range config.SignalGroup {
		for _, source := range group.AzureResource {
			for i := range source.Signal {
				if !configured("azure_resource", i) {
					source.Signal[i].RefreshInterval = ""
				}
			}
		}
		for _, source := range group.AzureLogAnalytics {
			for i := range source.Signal {
				if !configured("azure_log_analytics", i) {
					source.Signal[i].RefreshInterval = ""
				}
			}
		}
		for _, source := range group.AzureMonitorWorkspace {
			for i := range source.Signal {
				if !configured("azure_monitor_workspace", i) {
					source.Signal[i].RefreshInterval = ""
				}
			}
		}
	}
}

func mergeOptional[T any](current, planned *T, path string, hasChange func(string) bool, merge func(T, T) T) *T {
	if current == nil || planned == nil {
		return changedValue(hasChange, path, current, planned)
	}
	result := merge(*current, *planned)
	return &result
}

func mergeHealthModelEntityProperties(current entities.EntityProperties, planned *entities.EntityProperties, hasChange func(string) bool) *entities.EntityProperties {
	current.ProvisioningState, current.DiscoveredBy, current.HealthState = nil, nil, nil
	current.DisplayName = changedValue(hasChange, "display_name", current.DisplayName, planned.DisplayName)
	current.HealthObjective = changedValue(hasChange, "health_objective_percentage", current.HealthObjective, planned.HealthObjective)
	current.Impact = changedValue(hasChange, "impact", current.Impact, planned.Impact)
	current.CanvasPosition = mergeOptional(current.CanvasPosition, planned.CanvasPosition, "canvas_position", hasChange, func(old, next entities.EntityCoordinates) entities.EntityCoordinates {
		old.X = changedValue(hasChange, "canvas_position.0.x", old.X, next.X)
		old.Y = changedValue(hasChange, "canvas_position.0.y", old.Y, next.Y)
		return old
	})
	current.Icon = mergeOptional(current.Icon, planned.Icon, "icon", hasChange, func(old, next entities.IconDefinition) entities.IconDefinition {
		old.IconName = changedValue(hasChange, "icon.0.name", old.IconName, next.IconName)
		old.CustomData = changedValue(hasChange, "icon.0.custom_data_base64", old.CustomData, next.CustomData)
		return old
	})
	current.Alerts = mergeOptional(current.Alerts, planned.Alerts, "alert", hasChange, func(old, next entities.EntityAlerts) entities.EntityAlerts {
		old.Unhealthy = mergeHealthModelEntityAlert(old.Unhealthy, next.Unhealthy, "alert.0.unhealthy", hasChange)
		old.Degraded = mergeHealthModelEntityAlert(old.Degraded, next.Degraded, "alert.0.degraded", hasChange)
		return old
	})
	current.Tags = changedValue(hasChange, "tags", current.Tags, planned.Tags)
	current.SignalGroups = mergeHealthModelEntitySignalGroups(current.SignalGroups, planned.SignalGroups, hasChange)
	if current.SignalAggregationGroups != nil {
		groups := append([]entities.SignalAggregationGroup{}, (*current.SignalAggregationGroups)...)
		for i := range groups {
			groups[i].AggregatedHealthState, groups[i].UnresolvedMembers = nil, nil
		}
		current.SignalAggregationGroups = &groups
	}
	return &current
}

func mergeHealthModelEntitySignalGroups(current, planned *entities.SignalGroups, hasChange func(string) bool) *entities.SignalGroups {
	if current == nil || planned == nil {
		current = changedValue(hasChange, "signal_group", current, planned)
		if current == nil {
			return nil
		}
		planned = current
	}
	result := *current
	result.External = nil
	resourcePath := "signal_group.0.azure_resource"
	if current.AzureResource == nil || planned.AzureResource == nil {
		result.AzureResource = changedValue(hasChange, resourcePath, current.AzureResource, planned.AzureResource)
	} else {
		group := *current.AzureResource
		group.AuthenticationSetting = changedValue(hasChange, resourcePath+".0.authentication_setting_name", group.AuthenticationSetting, planned.AzureResource.AuthenticationSetting)
		group.AzureResourceId = changedValue(hasChange, resourcePath+".0.target_resource_id", group.AzureResourceId, planned.AzureResource.AzureResourceId)
		group.ResourceHealth = changedValue(hasChange, resourcePath+".0.resource_health_enabled", group.ResourceHealth, planned.AzureResource.ResourceHealth)
		group.Signals = mergeHealthModelEntityMetricSignals(group.Signals, planned.AzureResource.Signals, resourcePath+".0.signal", hasChange)
		result.AzureResource = &group
	}
	if result.AzureResource != nil {
		group := *result.AzureResource
		if group.ResourceHealth != nil {
			health := *group.ResourceHealth
			health.SignalName, health.Status = nil, nil
			group.ResourceHealth = &health
		}
		signals := append([]entities.AzureResourceSignal{}, pointer.From(group.Signals)...)
		for i := range signals {
			signals[i].Status = nil
		}
		if group.Signals != nil {
			group.Signals = &signals
		}
		result.AzureResource = &group
	}
	logPath := "signal_group.0.azure_log_analytics"
	if current.AzureLogAnalytics == nil || planned.AzureLogAnalytics == nil {
		result.AzureLogAnalytics = changedValue(hasChange, logPath, current.AzureLogAnalytics, planned.AzureLogAnalytics)
	} else {
		group := *current.AzureLogAnalytics
		group.AuthenticationSetting = changedValue(hasChange, logPath+".0.authentication_setting_name", group.AuthenticationSetting, planned.AzureLogAnalytics.AuthenticationSetting)
		group.LogAnalyticsWorkspaceResourceId = changedValue(hasChange, logPath+".0.log_analytics_workspace_id", group.LogAnalyticsWorkspaceResourceId, planned.AzureLogAnalytics.LogAnalyticsWorkspaceResourceId)
		group.Signals = mergeHealthModelEntityLogSignals(group.Signals, planned.AzureLogAnalytics.Signals, logPath+".0.signal", hasChange)
		result.AzureLogAnalytics = &group
	}
	if result.AzureLogAnalytics != nil {
		group := *result.AzureLogAnalytics
		signals := append([]entities.LogAnalyticsSignal{}, pointer.From(group.Signals)...)
		for i := range signals {
			signals[i].Status = nil
		}
		if group.Signals != nil {
			group.Signals = &signals
		}
		result.AzureLogAnalytics = &group
	}
	prometheusPath := "signal_group.0.azure_monitor_workspace"
	if current.AzureMonitorWorkspace == nil || planned.AzureMonitorWorkspace == nil {
		result.AzureMonitorWorkspace = changedValue(hasChange, prometheusPath, current.AzureMonitorWorkspace, planned.AzureMonitorWorkspace)
	} else {
		group := *current.AzureMonitorWorkspace
		group.AuthenticationSetting = changedValue(hasChange, prometheusPath+".0.authentication_setting_name", group.AuthenticationSetting, planned.AzureMonitorWorkspace.AuthenticationSetting)
		group.AzureMonitorWorkspaceResourceId = changedValue(hasChange, prometheusPath+".0.monitor_workspace_id", group.AzureMonitorWorkspaceResourceId, planned.AzureMonitorWorkspace.AzureMonitorWorkspaceResourceId)
		group.Signals = mergeHealthModelEntityPrometheusSignals(group.Signals, planned.AzureMonitorWorkspace.Signals, prometheusPath+".0.signal", hasChange)
		result.AzureMonitorWorkspace = &group
	}
	if result.AzureMonitorWorkspace != nil {
		group := *result.AzureMonitorWorkspace
		signals := append([]entities.PrometheusMetricsSignal{}, pointer.From(group.Signals)...)
		for i := range signals {
			signals[i].Status = nil
		}
		if group.Signals != nil {
			group.Signals = &signals
		}
		result.AzureMonitorWorkspace = &group
	}
	dependencyPath := "signal_group.0.dependency"
	if current.Dependencies == nil || planned.Dependencies == nil {
		result.Dependencies = changedValue(hasChange, dependencyPath, current.Dependencies, planned.Dependencies)
	} else {
		dependency := *current.Dependencies
		dependency.AggregationType = changedValue(hasChange, dependencyPath+".0.aggregation_type", dependency.AggregationType, planned.Dependencies.AggregationType)
		dependency.Unit = changedValue(hasChange, dependencyPath+".0.unit", dependency.Unit, planned.Dependencies.Unit)
		dependency.UnhealthyThreshold = changedValue(hasChange, dependencyPath+".0.unhealthy_threshold", dependency.UnhealthyThreshold, planned.Dependencies.UnhealthyThreshold)
		dependency.DegradedThreshold = changedValue(hasChange, dependencyPath+".0.degraded_threshold", dependency.DegradedThreshold, planned.Dependencies.DegradedThreshold)
		dependency.IgnoreUnknown = changedValue(hasChange, dependencyPath+".0.ignore_unknown_enabled", dependency.IgnoreUnknown, planned.Dependencies.IgnoreUnknown)
		result.Dependencies = &dependency
	}
	return &result
}

func mergeHealthModelEntityMetricSignals(current, planned *[]entities.AzureResourceSignal, path string, hasChange func(string) bool) *[]entities.AzureResourceSignal {
	if !hasChange(path) {
		return current
	}
	if planned == nil {
		return nil
	}
	result := append([]entities.AzureResourceSignal{}, (*planned)...)
	for i, signal := range result {
		if current == nil || i >= len(*current) || (*current)[i].Name != signal.Name {
			continue
		}
		old := (*current)[i]
		prefix := fmt.Sprintf("%s.%d.", path, i)
		signal.MetricNamespace = changedValue(hasChange, prefix+"metric_namespace", old.MetricNamespace, signal.MetricNamespace)
		signal.MetricName = changedValue(hasChange, prefix+"metric_name", old.MetricName, signal.MetricName)
		signal.AggregationType = changedValue(hasChange, prefix+"aggregation_type", old.AggregationType, signal.AggregationType)
		signal.TimeGrain = changedValue(hasChange, prefix+"time_grain", old.TimeGrain, signal.TimeGrain)
		signal.DimensionFilter = changedValue(hasChange, prefix+"dimension_filter", old.DimensionFilter, signal.DimensionFilter)
		signal.DisplayName = changedValue(hasChange, prefix+"display_name", old.DisplayName, signal.DisplayName)
		signal.RefreshInterval = changedValue(hasChange, prefix+"refresh_interval", old.RefreshInterval, signal.RefreshInterval)
		signal.DataUnit = changedValue(hasChange, prefix+"data_unit", old.DataUnit, signal.DataUnit)
		signal.SignalDefinitionName = changedValue(hasChange, prefix+"signal_definition_name", old.SignalDefinitionName, signal.SignalDefinitionName)
		signal.EvaluationRules = mergeHealthModelEntityEvaluation(old.EvaluationRules, signal.EvaluationRules, prefix+"evaluation_rule", hasChange)
		result[i] = signal
	}
	return &result
}

func mergeHealthModelEntityLogSignals(current, planned *[]entities.LogAnalyticsSignal, path string, hasChange func(string) bool) *[]entities.LogAnalyticsSignal {
	if !hasChange(path) {
		return current
	}
	if planned == nil {
		return nil
	}
	result := append([]entities.LogAnalyticsSignal{}, (*planned)...)
	for i, signal := range result {
		if current == nil || i >= len(*current) || (*current)[i].Name != signal.Name {
			continue
		}
		old := (*current)[i]
		prefix := fmt.Sprintf("%s.%d.", path, i)
		signal.QueryText = changedValue(hasChange, prefix+"query_text", old.QueryText, signal.QueryText)
		signal.TimeGrain = changedValue(hasChange, prefix+"time_grain", old.TimeGrain, signal.TimeGrain)
		signal.ValueColumnName = changedValue(hasChange, prefix+"value_column_name", old.ValueColumnName, signal.ValueColumnName)
		signal.DisplayName = changedValue(hasChange, prefix+"display_name", old.DisplayName, signal.DisplayName)
		signal.RefreshInterval = changedValue(hasChange, prefix+"refresh_interval", old.RefreshInterval, signal.RefreshInterval)
		signal.DataUnit = changedValue(hasChange, prefix+"data_unit", old.DataUnit, signal.DataUnit)
		signal.SignalDefinitionName = changedValue(hasChange, prefix+"signal_definition_name", old.SignalDefinitionName, signal.SignalDefinitionName)
		signal.EvaluationRules = mergeHealthModelEntityEvaluation(old.EvaluationRules, signal.EvaluationRules, prefix+"evaluation_rule", hasChange)
		result[i] = signal
	}
	return &result
}

func mergeHealthModelEntityPrometheusSignals(current, planned *[]entities.PrometheusMetricsSignal, path string, hasChange func(string) bool) *[]entities.PrometheusMetricsSignal {
	if !hasChange(path) {
		return current
	}
	if planned == nil {
		return nil
	}
	result := append([]entities.PrometheusMetricsSignal{}, (*planned)...)
	for i, signal := range result {
		if current == nil || i >= len(*current) || (*current)[i].Name != signal.Name {
			continue
		}
		old := (*current)[i]
		prefix := fmt.Sprintf("%s.%d.", path, i)
		signal.QueryText = changedValue(hasChange, prefix+"query_text", old.QueryText, signal.QueryText)
		signal.TimeGrain = changedValue(hasChange, prefix+"time_grain", old.TimeGrain, signal.TimeGrain)
		signal.DisplayName = changedValue(hasChange, prefix+"display_name", old.DisplayName, signal.DisplayName)
		signal.RefreshInterval = changedValue(hasChange, prefix+"refresh_interval", old.RefreshInterval, signal.RefreshInterval)
		signal.DataUnit = changedValue(hasChange, prefix+"data_unit", old.DataUnit, signal.DataUnit)
		signal.SignalDefinitionName = changedValue(hasChange, prefix+"signal_definition_name", old.SignalDefinitionName, signal.SignalDefinitionName)
		signal.EvaluationRules = mergeHealthModelEntityEvaluation(old.EvaluationRules, signal.EvaluationRules, prefix+"evaluation_rule", hasChange)
		result[i] = signal
	}
	return &result
}

func mergeHealthModelEntityAlert(current, planned *entities.AlertConfiguration, path string, hasChange func(string) bool) *entities.AlertConfiguration {
	return mergeOptional(current, planned, path, hasChange, func(old, next entities.AlertConfiguration) entities.AlertConfiguration {
		old.Severity = changedValue(hasChange, path+".0.severity", old.Severity, next.Severity)
		old.Description = changedValue(hasChange, path+".0.description", old.Description, next.Description)
		old.ActionGroupIds = changedValue(hasChange, path+".0.action_group_ids", old.ActionGroupIds, next.ActionGroupIds)
		return old
	})
}

func mergeHealthModelEntityEvaluation(current, planned *entities.EvaluationRule, path string, hasChange func(string) bool) *entities.EvaluationRule {
	mergeThreshold := func(old, next entities.ThresholdRuleV2, path string) entities.ThresholdRuleV2 {
		old.Operator = changedValue(hasChange, path+".0.operator", old.Operator, next.Operator)
		old.Threshold = changedValue(hasChange, path+".0.threshold", old.Threshold, next.Threshold)
		old.Sensitivity = changedValue(hasChange, path+".0.sensitivity", old.Sensitivity, next.Sensitivity)
		return old
	}
	return mergeOptional(current, planned, path, hasChange, func(old, next entities.EvaluationRule) entities.EvaluationRule {
		old.UnhealthyRule = mergeThreshold(old.UnhealthyRule, next.UnhealthyRule, path+".0.unhealthy")
		old.DegradedRule = mergeOptional(old.DegradedRule, next.DegradedRule, path+".0.degraded", hasChange, func(old, next entities.ThresholdRuleV2) entities.ThresholdRuleV2 {
			return mergeThreshold(old, next, path+".0.degraded")
		})
		return old
	})
}

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

func flattenHealthModelEntitySignalGroups(input *entities.SignalGroups) ([]SignalGroupModel, error) {
	if input == nil {
		return []SignalGroupModel{}, nil
	}

	logAnalytics, err := flattenHealthModelEntityLogAnalyticsSignals(input.AzureLogAnalytics)
	if err != nil {
		return nil, err
	}
	monitorWorkspace, err := flattenHealthModelEntityAzureMonitorWorkspaceSignals(input.AzureMonitorWorkspace)
	if err != nil {
		return nil, err
	}
	output := SignalGroupModel{
		AzureResource:         flattenHealthModelEntityAzureResourceSignals(input.AzureResource),
		AzureLogAnalytics:     logAnalytics,
		AzureMonitorWorkspace: monitorWorkspace,
		Dependency:            flattenHealthModelEntityDependencies(input.Dependencies),
	}

	// The `external` signal group is populated by the service and has no Terraform representation,
	// so a response that carries nothing else must still flatten to an empty block rather than one
	// holding four empty lists.
	if len(output.AzureResource) == 0 && len(output.AzureLogAnalytics) == 0 && len(output.AzureMonitorWorkspace) == 0 && len(output.Dependency) == 0 {
		return []SignalGroupModel{}, nil
	}

	return []SignalGroupModel{output}, nil
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
			EvaluationRules: expandEvaluationRule(signal.EvaluationRule),
		}
		if signal.MetricNamespace != "" {
			item.MetricNamespace = pointer.To(signal.MetricNamespace)
		}
		if signal.MetricName != "" {
			item.MetricName = pointer.To(signal.MetricName)
		}
		if signal.AggregationType != "" {
			item.AggregationType = pointer.To(entities.MetricAggregationType(signal.AggregationType))
		}
		if signal.TimeGrain != "" {
			item.TimeGrain = pointer.To(signal.TimeGrain)
		}

		if signal.DisplayName != "" {
			item.DisplayName = pointer.To(signal.DisplayName)
		}
		if signal.DimensionFilter != "" {
			item.DimensionFilter = pointer.To(signal.DimensionFilter)
		}
		if signal.RefreshInterval != "" {
			item.RefreshInterval = pointer.To(entities.RefreshInterval(signal.RefreshInterval))
		} else if signal.SignalDefinitionName == "" {
			item.RefreshInterval = pointer.To(entities.RefreshIntervalPTOneM)
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
			EvaluationRules: expandEvaluationRule(signal.EvaluationRule),
		}
		if signal.QueryText != "" {
			item.QueryText = pointer.To(signal.QueryText)
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
		} else if signal.SignalDefinitionName == "" {
			item.RefreshInterval = pointer.To(entities.RefreshIntervalPTOneM)
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

func flattenHealthModelEntityLogAnalyticsSignals(input *entities.LogAnalyticsSignals) ([]LogAnalyticsSignalGroupModel, error) {
	if input == nil {
		return []LogAnalyticsSignalGroupModel{}, nil
	}

	workspace, err := workspaces.ParseWorkspaceIDInsensitively(input.LogAnalyticsWorkspaceResourceId)
	if err != nil {
		return nil, fmt.Errorf("parsing `log_analytics_workspace_id`: %+v", err)
	}
	output := LogAnalyticsSignalGroupModel{
		AuthenticationSettingName: input.AuthenticationSetting,
		LogAnalyticsWorkspaceId:   workspace.ID(),
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

	return []LogAnalyticsSignalGroupModel{output}, nil
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
			EvaluationRules: expandEvaluationRule(signal.EvaluationRule),
		}
		if signal.QueryText != "" {
			item.QueryText = pointer.To(signal.QueryText)
		}

		if signal.TimeGrain != "" {
			item.TimeGrain = pointer.To(signal.TimeGrain)
		}
		if signal.DisplayName != "" {
			item.DisplayName = pointer.To(signal.DisplayName)
		}
		if signal.RefreshInterval != "" {
			item.RefreshInterval = pointer.To(entities.RefreshInterval(signal.RefreshInterval))
		} else if signal.SignalDefinitionName == "" {
			item.RefreshInterval = pointer.To(entities.RefreshIntervalPTOneM)
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

func flattenHealthModelEntityAzureMonitorWorkspaceSignals(input *entities.AzureMonitorWorkspaceSignals) ([]AzureMonitorWorkspaceSignalGroupModel, error) {
	if input == nil {
		return []AzureMonitorWorkspaceSignalGroupModel{}, nil
	}

	workspace, err := azuremonitorworkspaces.ParseAccountIDInsensitively(input.AzureMonitorWorkspaceResourceId)
	if err != nil {
		return nil, fmt.Errorf("parsing `monitor_workspace_id`: %+v", err)
	}
	output := AzureMonitorWorkspaceSignalGroupModel{
		AuthenticationSettingName: input.AuthenticationSetting,
		MonitorWorkspaceId:        workspace.ID(),
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

	return []AzureMonitorWorkspaceSignalGroupModel{output}, nil
}

func expandHealthModelEntityDependencies(input []DependencySignalGroupModel, degradedThresholdConfigured bool) *entities.DependenciesSignalGroupV2 {
	if len(input) == 0 {
		return nil
	}

	config := input[0]

	output := entities.DependenciesSignalGroupV2{
		AggregationType: entities.AggregationType(config.AggregationType),
		IgnoreUnknown:   pointer.To(config.IgnoreUnknownEnabled),
	}

	if config.Unit != "" {
		output.Unit = pointer.To(entities.AggregationUnit(config.Unit))
	}

	// `unhealthy_threshold` is `TypeFloat`, so a decoded `0` is indistinguishable from an absent
	// one - and the service both accepts and stores a threshold of `0`. Keying off the aggregation
	// type instead recovers the distinction: the plan-time guard has already established that the
	// counting aggregations always carry a threshold and that `WorstOf` and `BestOf` never do, so the
	// aggregation type says which case this is.
	if config.AggregationType != string(entities.AggregationTypeWorstOf) && config.AggregationType != string(entities.AggregationTypeBestOf) {
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

func flattenHealthModelEntityAlerts(input *entities.EntityAlerts) ([]EntityAlertModel, error) {
	if input == nil || input.Unhealthy == nil && input.Degraded == nil {
		return []EntityAlertModel{}, nil
	}

	unhealthy, err := flattenHealthModelEntityAlertConfiguration(input.Unhealthy)
	if err != nil {
		return nil, fmt.Errorf("flattening `alert.unhealthy`: %+v", err)
	}
	degraded, err := flattenHealthModelEntityAlertConfiguration(input.Degraded)
	if err != nil {
		return nil, fmt.Errorf("flattening `alert.degraded`: %+v", err)
	}
	return []EntityAlertModel{{
		Unhealthy: unhealthy,
		Degraded:  degraded,
	}}, nil
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

func flattenHealthModelEntityAlertConfiguration(input *entities.AlertConfiguration) ([]AlertConfigurationModel, error) {
	if input == nil {
		return []AlertConfigurationModel{}, nil
	}

	actionGroups := make([]string, 0, len(pointer.From(input.ActionGroupIds)))
	for _, value := range pointer.From(input.ActionGroupIds) {
		id, err := actiongroupsapis.ParseActionGroupIDInsensitively(value)
		if err != nil {
			return nil, fmt.Errorf("parsing `action_group_ids`: %+v", err)
		}
		actionGroups = append(actionGroups, id.ID())
	}
	return []AlertConfigurationModel{{
		Severity:       string(input.Severity),
		Description:    pointer.From(input.Description),
		ActionGroupIds: actionGroups,
	}}, nil
}
