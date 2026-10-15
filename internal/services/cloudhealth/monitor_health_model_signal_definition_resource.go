// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cloudhealth

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/terraform-provider-azurerm/internal/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/cloudhealth/sdk/2026-11-01/signaldefinitions"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/cloudhealth/validate"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/validation"
)

type HealthModelSignalDefinitionResource struct{}

var _ sdk.ResourceWithUpdate = HealthModelSignalDefinitionResource{}

var _ sdk.ResourceWithCustomizeDiff = HealthModelSignalDefinitionResource{}

type HealthModelSignalDefinitionResourceModel struct {
	Name                   string                           `tfschema:"name"`
	HealthModelId          string                           `tfschema:"health_model_id"`
	EvaluationRule         []EvaluationRuleModel            `tfschema:"evaluation_rule"`
	AzureResourceMetric    []AzureResourceMetricSignalModel `tfschema:"azure_resource_metric"`
	LogAnalyticsQuery      []LogAnalyticsQuerySignalModel   `tfschema:"log_analytics_query"`
	PrometheusMetricsQuery []PrometheusQuerySignalModel     `tfschema:"prometheus_metrics_query"`
	DisplayName            string                           `tfschema:"display_name"`
	RefreshInterval        string                           `tfschema:"refresh_interval"`
	DataUnit               string                           `tfschema:"data_unit"`
	Tags                   map[string]string                `tfschema:"tags"`
}

type AzureResourceMetricSignalModel struct {
	MetricNamespace string `tfschema:"metric_namespace"`
	MetricName      string `tfschema:"metric_name"`
	AggregationType string `tfschema:"aggregation_type"`
	TimeGrain       string `tfschema:"time_grain"`
	DimensionFilter string `tfschema:"dimension_filter"`
}

type LogAnalyticsQuerySignalModel struct {
	QueryText       string `tfschema:"query_text"`
	TimeGrain       string `tfschema:"time_grain"`
	ValueColumnName string `tfschema:"value_column_name"`
}

type PrometheusQuerySignalModel struct {
	QueryText string `tfschema:"query_text"`
	TimeGrain string `tfschema:"time_grain"`
}

func (HealthModelSignalDefinitionResource) ResourceType() string {
	return "azurerm_monitor_health_model_signal_definition"
}

func (HealthModelSignalDefinitionResource) ModelObject() interface{} {
	return &HealthModelSignalDefinitionResourceModel{}
}

func (HealthModelSignalDefinitionResource) IDValidationFunc() pluginsdk.SchemaValidateFunc {
	return signaldefinitions.ValidateSignaldefinitionID
}

func (HealthModelSignalDefinitionResource) Arguments() map[string]*pluginsdk.Schema {
	kinds := []string{"azure_resource_metric", "log_analytics_query", "prometheus_metrics_query"}

	return map[string]*pluginsdk.Schema{
		"name": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validate.HealthModelChildName,
		},

		"health_model_id": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: signaldefinitions.ValidateHealthmodelID,
		},

		"evaluation_rule": evaluationRuleSchema(),

		"azure_resource_metric": {
			Type:         pluginsdk.TypeList,
			Optional:     true,
			ForceNew:     true,
			MaxItems:     1,
			ExactlyOneOf: kinds,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"metric_namespace": {
						Type:         pluginsdk.TypeString,
						Required:     true,
						ValidateFunc: validation.StringLenBetween(1, 256),
					},

					"metric_name": {
						Type:         pluginsdk.TypeString,
						Required:     true,
						ValidateFunc: validation.StringLenBetween(1, 256),
					},

					"aggregation_type": {
						Type:         pluginsdk.TypeString,
						Required:     true,
						ValidateFunc: validation.StringInSlice(signaldefinitions.PossibleValuesForMetricAggregationType(), false),
					},

					"time_grain": {
						Type:         pluginsdk.TypeString,
						Required:     true,
						ValidateFunc: validation.All(validation.StringLenBetween(1, 100), validation.ISO8601Duration),
					},

					"dimension_filter": {
						Type:         pluginsdk.TypeString,
						Optional:     true,
						ValidateFunc: validation.StringLenBetween(1, 2048),
					},
				},
			},
		},

		"log_analytics_query": {
			Type:         pluginsdk.TypeList,
			Optional:     true,
			ForceNew:     true,
			MaxItems:     1,
			ExactlyOneOf: kinds,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"query_text": {
						Type:         pluginsdk.TypeString,
						Required:     true,
						ValidateFunc: validation.StringLenBetween(1, 5000),
					},

					"time_grain": {
						Type:         pluginsdk.TypeString,
						Optional:     true,
						ValidateFunc: validation.All(validation.StringLenBetween(1, 100), validation.ISO8601Duration),
					},

					"value_column_name": {
						Type:         pluginsdk.TypeString,
						Optional:     true,
						ValidateFunc: validation.StringLenBetween(1, 100),
					},
				},
			},
		},

		"prometheus_metrics_query": {
			Type:         pluginsdk.TypeList,
			Optional:     true,
			ForceNew:     true,
			MaxItems:     1,
			ExactlyOneOf: kinds,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"query_text": {
						Type:         pluginsdk.TypeString,
						Required:     true,
						ValidateFunc: validation.StringLenBetween(1, 5000),
					},

					"time_grain": {
						Type:         pluginsdk.TypeString,
						Optional:     true,
						ValidateFunc: validation.All(validation.StringLenBetween(1, 100), validation.ISO8601Duration),
					},
				},
			},
		},

		"display_name": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringLenBetween(1, 260),
		},

		"refresh_interval": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			Default:      string(signaldefinitions.RefreshIntervalPTOneM),
			ValidateFunc: validation.StringInSlice(signaldefinitions.PossibleValuesForRefreshInterval(), false),
		},

		"data_unit": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringLenBetween(1, 100),
		},

		"tags": commonschema.Tags(),
	}
}

func (HealthModelSignalDefinitionResource) Attributes() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{}
}

func (r HealthModelSignalDefinitionResource) CustomizeDiff() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 5 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			return validateEvaluationRuleRawConfig(metadata.ResourceDiff.GetRawConfig())
		},
	}
}

func (r HealthModelSignalDefinitionResource) Create() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.CloudHealth.SignalDefinitions

			var config HealthModelSignalDefinitionResourceModel
			if err := metadata.Decode(&config); err != nil {
				return fmt.Errorf("decoding: %+v", err)
			}

			healthModelId, err := signaldefinitions.ParseHealthmodelID(config.HealthModelId)
			if err != nil {
				return err
			}

			id := signaldefinitions.NewSignaldefinitionID(healthModelId.SubscriptionId, healthModelId.ResourceGroupName, healthModelId.HealthmodelName, config.Name)

			if !metadata.Client.Features.SkipImportCheckOnCreateAndAllowOverwritingExistingResources {
				existing, err := client.Get(ctx, id)
				if err != nil && !response.WasNotFound(existing.HttpResponse) {
					return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
				}
				if !response.WasNotFound(existing.HttpResponse) {
					return metadata.ResourceRequiresImport(r.ResourceType(), id)
				}
			}

			props, err := expandHealthModelSignalDefinitionProperties(config)
			if err != nil {
				return err
			}

			payload := signaldefinitions.SignalDefinition{
				Properties: props,
			}

			if err := client.CreateOrUpdateThenPoll(ctx, id, payload); err != nil {
				return fmt.Errorf("creating %s: %+v", id, err)
			}

			metadata.SetID(id)

			return nil
		},
	}
}

func (HealthModelSignalDefinitionResource) Read() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 5 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.CloudHealth.SignalDefinitions

			id, err := signaldefinitions.ParseSignaldefinitionID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			resp, err := client.Get(ctx, *id)
			if err != nil {
				if response.WasNotFound(resp.HttpResponse) {
					return metadata.MarkAsGone(id)
				}
				return fmt.Errorf("retrieving %s: %+v", *id, err)
			}

			state := HealthModelSignalDefinitionResourceModel{
				Name:          id.SignaldefinitionName,
				HealthModelId: signaldefinitions.NewHealthmodelID(id.SubscriptionId, id.ResourceGroupName, id.HealthmodelName).ID(),
			}

			if model := resp.Model; model != nil {
				if err := flattenHealthModelSignalDefinitionProperties(model.Properties, &state); err != nil {
					return fmt.Errorf("flattening properties for %s: %+v", *id, err)
				}
			}

			return metadata.Encode(&state)
		},
	}
}

func (HealthModelSignalDefinitionResource) Update() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.CloudHealth.SignalDefinitions

			id, err := signaldefinitions.ParseSignaldefinitionID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			var config HealthModelSignalDefinitionResourceModel
			if err := metadata.Decode(&config); err != nil {
				return fmt.Errorf("decoding: %+v", err)
			}

			props, err := expandHealthModelSignalDefinitionProperties(config)
			if err != nil {
				return err
			}

			existing, err := client.Get(ctx, *id)
			if err != nil {
				return fmt.Errorf("retrieving %s: %+v", *id, err)
			}
			if existing.Model == nil || existing.Model.Properties == nil {
				return fmt.Errorf("retrieving %s: missing model or properties", *id)
			}
			props, err = mergeHealthModelSignalDefinitionProperties(existing.Model.Properties, props, cloudHealthChanges(metadata.ResourceData))
			if err != nil {
				return fmt.Errorf("merging properties for %s: %+v", *id, err)
			}
			payload := signaldefinitions.SignalDefinition{
				Properties: props,
			}

			if err := client.CreateOrUpdateThenPoll(ctx, *id, payload); err != nil {
				return fmt.Errorf("updating %s: %+v", *id, err)
			}

			return nil
		},
	}
}

func (HealthModelSignalDefinitionResource) Delete() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.CloudHealth.SignalDefinitions

			id, err := signaldefinitions.ParseSignaldefinitionID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			if err := client.DeleteThenPoll(ctx, *id); err != nil {
				return fmt.Errorf("deleting %s: %+v", *id, err)
			}

			return nil
		},
	}
}

func mergeHealthModelSignalDefinitionProperties(current, planned signaldefinitions.SignalDefinitionProperties, hasChange func(string) bool) (signaldefinitions.SignalDefinitionProperties, error) {
	switch props := current.(type) {
	case signaldefinitions.ResourceMetricSignalDefinitionProperties:
		next, ok := planned.(signaldefinitions.ResourceMetricSignalDefinitionProperties)
		if !ok {
			return nil, fmt.Errorf("signal kind changed")
		}
		props.ProvisioningState = nil
		props.DisplayName = changedValue(hasChange, "display_name", props.DisplayName, next.DisplayName)
		props.DataUnit = changedValue(hasChange, "data_unit", props.DataUnit, next.DataUnit)
		props.RefreshInterval = changedValue(hasChange, "refresh_interval", props.RefreshInterval, next.RefreshInterval)
		props.Tags = changedValue(hasChange, "tags", props.Tags, next.Tags)
		props.EvaluationRules = mergeSignalDefinitionEvaluationRule(props.EvaluationRules, next.EvaluationRules, hasChange)
		props.MetricNamespace = changedValue(hasChange, "azure_resource_metric.0.metric_namespace", props.MetricNamespace, next.MetricNamespace)
		props.MetricName = changedValue(hasChange, "azure_resource_metric.0.metric_name", props.MetricName, next.MetricName)
		props.AggregationType = changedValue(hasChange, "azure_resource_metric.0.aggregation_type", props.AggregationType, next.AggregationType)
		props.TimeGrain = changedValue(hasChange, "azure_resource_metric.0.time_grain", props.TimeGrain, next.TimeGrain)
		props.DimensionFilter = changedValue(hasChange, "azure_resource_metric.0.dimension_filter", props.DimensionFilter, next.DimensionFilter)
		return props, nil
	case signaldefinitions.LogAnalyticsQuerySignalDefinitionProperties:
		next, ok := planned.(signaldefinitions.LogAnalyticsQuerySignalDefinitionProperties)
		if !ok {
			return nil, fmt.Errorf("signal kind changed")
		}
		props.ProvisioningState = nil
		props.DisplayName = changedValue(hasChange, "display_name", props.DisplayName, next.DisplayName)
		props.DataUnit = changedValue(hasChange, "data_unit", props.DataUnit, next.DataUnit)
		props.RefreshInterval = changedValue(hasChange, "refresh_interval", props.RefreshInterval, next.RefreshInterval)
		props.Tags = changedValue(hasChange, "tags", props.Tags, next.Tags)
		props.EvaluationRules = mergeSignalDefinitionEvaluationRule(props.EvaluationRules, next.EvaluationRules, hasChange)
		props.QueryText = changedValue(hasChange, "log_analytics_query.0.query_text", props.QueryText, next.QueryText)
		props.TimeGrain = changedValue(hasChange, "log_analytics_query.0.time_grain", props.TimeGrain, next.TimeGrain)
		props.ValueColumnName = changedValue(hasChange, "log_analytics_query.0.value_column_name", props.ValueColumnName, next.ValueColumnName)
		return props, nil
	case signaldefinitions.PrometheusMetricsSignalDefinitionProperties:
		next, ok := planned.(signaldefinitions.PrometheusMetricsSignalDefinitionProperties)
		if !ok {
			return nil, fmt.Errorf("signal kind changed")
		}
		props.ProvisioningState = nil
		props.DisplayName = changedValue(hasChange, "display_name", props.DisplayName, next.DisplayName)
		props.DataUnit = changedValue(hasChange, "data_unit", props.DataUnit, next.DataUnit)
		props.RefreshInterval = changedValue(hasChange, "refresh_interval", props.RefreshInterval, next.RefreshInterval)
		props.Tags = changedValue(hasChange, "tags", props.Tags, next.Tags)
		props.EvaluationRules = mergeSignalDefinitionEvaluationRule(props.EvaluationRules, next.EvaluationRules, hasChange)
		props.QueryText = changedValue(hasChange, "prometheus_metrics_query.0.query_text", props.QueryText, next.QueryText)
		props.TimeGrain = changedValue(hasChange, "prometheus_metrics_query.0.time_grain", props.TimeGrain, next.TimeGrain)
		return props, nil
	default:
		return nil, fmt.Errorf("unhandled signal kind %q", current.SignalDefinitionProperties().SignalKind)
	}
}

func expandHealthModelSignalDefinitionProperties(config HealthModelSignalDefinitionResourceModel) (signaldefinitions.SignalDefinitionProperties, error) {
	base := signaldefinitions.BaseSignalDefinitionPropertiesImpl{
		EvaluationRules: expandSignalDefinitionEvaluationRule(config.EvaluationRule),
		Tags:            pointer.To(config.Tags),
	}

	if config.DisplayName != "" {
		base.DisplayName = pointer.To(config.DisplayName)
	}
	if config.RefreshInterval != "" {
		base.RefreshInterval = pointer.To(signaldefinitions.RefreshInterval(config.RefreshInterval))
	}
	if config.DataUnit != "" {
		base.DataUnit = pointer.To(config.DataUnit)
	}

	if len(config.AzureResourceMetric) > 0 {
		metric := config.AzureResourceMetric[0]
		output := signaldefinitions.ResourceMetricSignalDefinitionProperties{
			SignalKind:      signaldefinitions.SignalKindAzureResourceMetric,
			MetricNamespace: metric.MetricNamespace,
			MetricName:      metric.MetricName,
			AggregationType: signaldefinitions.MetricAggregationType(metric.AggregationType),
			TimeGrain:       metric.TimeGrain,
			DataUnit:        base.DataUnit,
			DisplayName:     base.DisplayName,
			EvaluationRules: base.EvaluationRules,
			RefreshInterval: base.RefreshInterval,
			Tags:            base.Tags,
		}
		if metric.DimensionFilter != "" {
			output.DimensionFilter = pointer.To(metric.DimensionFilter)
		}
		return output, nil
	}

	if len(config.LogAnalyticsQuery) > 0 {
		query := config.LogAnalyticsQuery[0]
		output := signaldefinitions.LogAnalyticsQuerySignalDefinitionProperties{
			SignalKind:      signaldefinitions.SignalKindLogAnalyticsQuery,
			QueryText:       query.QueryText,
			DataUnit:        base.DataUnit,
			DisplayName:     base.DisplayName,
			EvaluationRules: base.EvaluationRules,
			RefreshInterval: base.RefreshInterval,
			Tags:            base.Tags,
		}
		if query.TimeGrain != "" {
			output.TimeGrain = pointer.To(query.TimeGrain)
		}
		if query.ValueColumnName != "" {
			output.ValueColumnName = pointer.To(query.ValueColumnName)
		}
		return output, nil
	}

	if len(config.PrometheusMetricsQuery) == 0 {
		return nil, fmt.Errorf("exactly one of `azure_resource_metric`, `log_analytics_query` or `prometheus_metrics_query` must be set")
	}

	query := config.PrometheusMetricsQuery[0]
	output := signaldefinitions.PrometheusMetricsSignalDefinitionProperties{
		SignalKind:      signaldefinitions.SignalKindPrometheusMetricsQuery,
		QueryText:       query.QueryText,
		DataUnit:        base.DataUnit,
		DisplayName:     base.DisplayName,
		EvaluationRules: base.EvaluationRules,
		RefreshInterval: base.RefreshInterval,
		Tags:            base.Tags,
	}
	if query.TimeGrain != "" {
		output.TimeGrain = pointer.To(query.TimeGrain)
	}
	return output, nil
}

func flattenHealthModelSignalDefinitionProperties(input signaldefinitions.SignalDefinitionProperties, state *HealthModelSignalDefinitionResourceModel) error {
	if input == nil {
		return nil
	}

	base := input.SignalDefinitionProperties()

	state.EvaluationRule = flattenSignalDefinitionEvaluationRule(base.EvaluationRules)
	state.DisplayName = pointer.From(base.DisplayName)
	state.DataUnit = pointer.From(base.DataUnit)
	state.Tags = pointer.From(base.Tags)

	if base.RefreshInterval != nil {
		state.RefreshInterval = string(*base.RefreshInterval)
	}

	switch props := input.(type) {
	case signaldefinitions.ResourceMetricSignalDefinitionProperties:
		state.AzureResourceMetric = []AzureResourceMetricSignalModel{{
			MetricNamespace: props.MetricNamespace,
			MetricName:      props.MetricName,
			AggregationType: string(props.AggregationType),
			TimeGrain:       props.TimeGrain,
			DimensionFilter: pointer.From(props.DimensionFilter),
		}}
	case signaldefinitions.LogAnalyticsQuerySignalDefinitionProperties:
		state.LogAnalyticsQuery = []LogAnalyticsQuerySignalModel{{
			QueryText:       props.QueryText,
			TimeGrain:       pointer.From(props.TimeGrain),
			ValueColumnName: pointer.From(props.ValueColumnName),
		}}
	case signaldefinitions.PrometheusMetricsSignalDefinitionProperties:
		state.PrometheusMetricsQuery = []PrometheusQuerySignalModel{{
			QueryText: props.QueryText,
			TimeGrain: pointer.From(props.TimeGrain),
		}}
	default:
		return fmt.Errorf("unhandled signal kind %q", base.SignalKind)
	}

	return nil
}

func mergeSignalDefinitionEvaluationRule(current, planned signaldefinitions.EvaluationRule, hasChange func(string) bool) signaldefinitions.EvaluationRule {
	mergeThreshold := func(old, next signaldefinitions.ThresholdRuleV2, path string) signaldefinitions.ThresholdRuleV2 {
		old.Operator = changedValue(hasChange, path+".0.operator", old.Operator, next.Operator)
		old.Threshold = changedValue(hasChange, path+".0.threshold", old.Threshold, next.Threshold)
		old.Sensitivity = changedValue(hasChange, path+".0.sensitivity", old.Sensitivity, next.Sensitivity)
		return old
	}
	current.UnhealthyRule = mergeThreshold(current.UnhealthyRule, planned.UnhealthyRule, "evaluation_rule.0.unhealthy")
	current.DegradedRule = mergeOptional(current.DegradedRule, planned.DegradedRule, "evaluation_rule.0.degraded", hasChange, func(old, next signaldefinitions.ThresholdRuleV2) signaldefinitions.ThresholdRuleV2 {
		return mergeThreshold(old, next, "evaluation_rule.0.degraded")
	})
	return current
}

func expandSignalDefinitionEvaluationRule(input []EvaluationRuleModel) signaldefinitions.EvaluationRule {
	if len(input) == 0 {
		return signaldefinitions.EvaluationRule{}
	}

	rules := input[0]
	output := signaldefinitions.EvaluationRule{}

	if len(rules.Unhealthy) > 0 {
		output.UnhealthyRule = expandSignalDefinitionThresholdRule(rules.Unhealthy[0])
	}

	if len(rules.Degraded) > 0 {
		degraded := expandSignalDefinitionThresholdRule(rules.Degraded[0])
		output.DegradedRule = &degraded
	}

	return output
}

func expandSignalDefinitionThresholdRule(input ThresholdRuleModel) signaldefinitions.ThresholdRuleV2 {
	output := signaldefinitions.ThresholdRuleV2{
		Operator: signaldefinitions.SignalOperator(input.Operator),
	}

	if output.Operator == signaldefinitions.SignalOperatorDynamic {
		if input.Sensitivity != "" {
			output.Sensitivity = pointer.To(signaldefinitions.DynamicThresholdSensitivity(input.Sensitivity))
		}
		return output
	}

	output.Threshold = pointer.To(input.Threshold)

	return output
}

func flattenSignalDefinitionEvaluationRule(input signaldefinitions.EvaluationRule) []EvaluationRuleModel {
	output := EvaluationRuleModel{
		Unhealthy: []ThresholdRuleModel{flattenSignalDefinitionThresholdRule(input.UnhealthyRule)},
	}

	if input.DegradedRule != nil {
		output.Degraded = []ThresholdRuleModel{flattenSignalDefinitionThresholdRule(*input.DegradedRule)}
	}

	return []EvaluationRuleModel{output}
}

func flattenSignalDefinitionThresholdRule(input signaldefinitions.ThresholdRuleV2) ThresholdRuleModel {
	output := ThresholdRuleModel{
		Operator:  string(input.Operator),
		Threshold: pointer.From(input.Threshold),
	}

	if input.Sensitivity != nil {
		output.Sensitivity = string(*input.Sensitivity)
	}

	return output
}
