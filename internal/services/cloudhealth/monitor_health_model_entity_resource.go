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
	"github.com/hashicorp/go-azure-sdk/resource-manager/insights/2023-01-01/actiongroupsapis"
	"github.com/hashicorp/go-azure-sdk/resource-manager/monitor/2023-04-03/azuremonitorworkspaces"
	"github.com/hashicorp/go-azure-sdk/resource-manager/operationalinsights/2020-08-01/workspaces"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/internal/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/cloudhealth/sdk/2026-11-01/entities"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/cloudhealth/validate"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/validation"
)

const healthModelEntityMaxSignals = 50

type HealthModelEntityResource struct{}

var _ sdk.ResourceWithUpdate = HealthModelEntityResource{}

var _ sdk.ResourceWithCustomizeDiff = HealthModelEntityResource{}

type HealthModelEntityResourceModel struct {
	Name                      string                `tfschema:"name"`
	HealthModelId             string                `tfschema:"health_model_id"`
	DisplayName               string                `tfschema:"display_name"`
	HealthObjectivePercentage *float64              `tfschema:"health_objective_percentage"`
	Impact                    string                `tfschema:"impact"`
	CanvasPosition            []CanvasPositionModel `tfschema:"canvas_position"`
	Icon                      []IconModel           `tfschema:"icon"`
	SignalGroup               []SignalGroupModel    `tfschema:"signal_group"`
	Alert                     []EntityAlertModel    `tfschema:"alert"`
	Tags                      map[string]string     `tfschema:"tags"`
	HealthState               string                `tfschema:"health_state"`
	DiscoveredBy              string                `tfschema:"discovered_by"`
}

type CanvasPositionModel struct {
	X float64 `tfschema:"x"`
	Y float64 `tfschema:"y"`
}

type IconModel struct {
	Name             string `tfschema:"name"`
	CustomDataBase64 string `tfschema:"custom_data_base64"`
}

type SignalGroupModel struct {
	AzureResource         []AzureResourceSignalGroupModel         `tfschema:"azure_resource"`
	AzureLogAnalytics     []LogAnalyticsSignalGroupModel          `tfschema:"azure_log_analytics"`
	AzureMonitorWorkspace []AzureMonitorWorkspaceSignalGroupModel `tfschema:"azure_monitor_workspace"`
	Dependency            []DependencySignalGroupModel            `tfschema:"dependency"`
}

type AzureResourceSignalGroupModel struct {
	AuthenticationSettingName string                     `tfschema:"authentication_setting_name"`
	TargetResourceId          string                     `tfschema:"target_resource_id"`
	ResourceHealthEnabled     bool                       `tfschema:"resource_health_enabled"`
	Signal                    []AzureResourceSignalModel `tfschema:"signal"`
}

type AzureResourceSignalModel struct {
	Name                 string                `tfschema:"name"`
	MetricNamespace      string                `tfschema:"metric_namespace"`
	MetricName           string                `tfschema:"metric_name"`
	AggregationType      string                `tfschema:"aggregation_type"`
	TimeGrain            string                `tfschema:"time_grain"`
	EvaluationRule       []EvaluationRuleModel `tfschema:"evaluation_rule"`
	DisplayName          string                `tfschema:"display_name"`
	DimensionFilter      string                `tfschema:"dimension_filter"`
	RefreshInterval      string                `tfschema:"refresh_interval"`
	DataUnit             string                `tfschema:"data_unit"`
	SignalDefinitionName string                `tfschema:"signal_definition_name"`
}

type LogAnalyticsSignalGroupModel struct {
	AuthenticationSettingName string                    `tfschema:"authentication_setting_name"`
	LogAnalyticsWorkspaceId   string                    `tfschema:"log_analytics_workspace_id"`
	Signal                    []LogAnalyticsSignalModel `tfschema:"signal"`
}

type LogAnalyticsSignalModel struct {
	Name                 string                `tfschema:"name"`
	QueryText            string                `tfschema:"query_text"`
	EvaluationRule       []EvaluationRuleModel `tfschema:"evaluation_rule"`
	TimeGrain            string                `tfschema:"time_grain"`
	ValueColumnName      string                `tfschema:"value_column_name"`
	DisplayName          string                `tfschema:"display_name"`
	RefreshInterval      string                `tfschema:"refresh_interval"`
	DataUnit             string                `tfschema:"data_unit"`
	SignalDefinitionName string                `tfschema:"signal_definition_name"`
}

type AzureMonitorWorkspaceSignalGroupModel struct {
	AuthenticationSettingName string                         `tfschema:"authentication_setting_name"`
	MonitorWorkspaceId        string                         `tfschema:"monitor_workspace_id"`
	Signal                    []PrometheusMetricsSignalModel `tfschema:"signal"`
}

type PrometheusMetricsSignalModel struct {
	Name                 string                `tfschema:"name"`
	QueryText            string                `tfschema:"query_text"`
	EvaluationRule       []EvaluationRuleModel `tfschema:"evaluation_rule"`
	TimeGrain            string                `tfschema:"time_grain"`
	DisplayName          string                `tfschema:"display_name"`
	RefreshInterval      string                `tfschema:"refresh_interval"`
	DataUnit             string                `tfschema:"data_unit"`
	SignalDefinitionName string                `tfschema:"signal_definition_name"`
}

type DependencySignalGroupModel struct {
	AggregationType      string  `tfschema:"aggregation_type"`
	Unit                 string  `tfschema:"unit"`
	UnhealthyThreshold   float64 `tfschema:"unhealthy_threshold"`
	DegradedThreshold    float64 `tfschema:"degraded_threshold"`
	IgnoreUnknownEnabled bool    `tfschema:"ignore_unknown_enabled"`
}

type EntityAlertModel struct {
	Unhealthy []AlertConfigurationModel `tfschema:"unhealthy"`
	Degraded  []AlertConfigurationModel `tfschema:"degraded"`
}

type AlertConfigurationModel struct {
	Severity       string   `tfschema:"severity"`
	Description    string   `tfschema:"description"`
	ActionGroupIds []string `tfschema:"action_group_ids"`
}

func (HealthModelEntityResource) ResourceType() string {
	return "azurerm_monitor_health_model_entity"
}

func (HealthModelEntityResource) ModelObject() interface{} {
	return &HealthModelEntityResourceModel{}
}

func (HealthModelEntityResource) IDValidationFunc() pluginsdk.SchemaValidateFunc {
	return entities.ValidateEntityID
}

func (HealthModelEntityResource) Arguments() map[string]*pluginsdk.Schema {
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
			ValidateFunc: entities.ValidateHealthmodelID,
		},

		"display_name": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringIsNotEmpty,
		},

		"health_objective_percentage": {
			Type:         pluginsdk.TypeFloat,
			Optional:     true,
			ValidateFunc: validation.FloatBetween(0, 100),
		},

		"impact": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			Default:      string(entities.EntityImpactStandard),
			ValidateFunc: validation.StringInSlice(validate.PossibleValuesForEntityImpact(), false),
		},

		"canvas_position": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"x": {
						Type:     pluginsdk.TypeFloat,
						Required: true,
					},

					"y": {
						Type:     pluginsdk.TypeFloat,
						Required: true,
					},
				},
			},
		},

		"icon": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"name": {
						Type:         pluginsdk.TypeString,
						Required:     true,
						ValidateFunc: validation.StringIsNotEmpty,
					},

					"custom_data_base64": {
						Type:         pluginsdk.TypeString,
						Optional:     true,
						ValidateFunc: validation.StringIsBase64,
					},
				},
			},
		},

		"signal_group": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"azure_resource": {
						Type:     pluginsdk.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &pluginsdk.Resource{
							Schema: map[string]*pluginsdk.Schema{
								"authentication_setting_name": {
									Type:         pluginsdk.TypeString,
									Required:     true,
									ValidateFunc: validate.HealthModelChildName,
								},

								"target_resource_id": {
									Type:         pluginsdk.TypeString,
									Required:     true,
									ValidateFunc: azure.ValidateResourceID,
								},

								"resource_health_enabled": {
									Type:     pluginsdk.TypeBool,
									Optional: true,
									Default:  true,
								},

								"signal": {
									Type:     pluginsdk.TypeList,
									Optional: true,
									MaxItems: healthModelEntityMaxSignals,
									Elem: &pluginsdk.Resource{
										Schema: map[string]*pluginsdk.Schema{
											"name": {
												Type:         pluginsdk.TypeString,
												Required:     true,
												ValidateFunc: validate.HealthModelChildName,
											},

											"metric_namespace": {
												Type:         pluginsdk.TypeString,
												Required:     true,
												ValidateFunc: validation.StringIsNotEmpty,
											},

											"metric_name": {
												Type:         pluginsdk.TypeString,
												Required:     true,
												ValidateFunc: validation.StringIsNotEmpty,
											},

											"aggregation_type": {
												Type:         pluginsdk.TypeString,
												Required:     true,
												ValidateFunc: validation.StringInSlice(validate.PossibleValuesForMetricAggregationType(), false),
											},

											"time_grain": {
												Type:         pluginsdk.TypeString,
												Required:     true,
												ValidateFunc: validation.ISO8601Duration,
											},

											"evaluation_rule": evaluationRuleSchema(),

											"display_name": {
												Type:         pluginsdk.TypeString,
												Optional:     true,
												ValidateFunc: validation.StringIsNotEmpty,
											},

											"dimension_filter": {
												Type:         pluginsdk.TypeString,
												Optional:     true,
												ValidateFunc: validation.StringIsNotEmpty,
											},

											"refresh_interval": {
												Type:         pluginsdk.TypeString,
												Optional:     true,
												Default:      string(entities.RefreshIntervalPTOneM),
												ValidateFunc: validation.StringInSlice(validate.PossibleValuesForRefreshInterval(), false),
											},

											"data_unit": {
												Type:         pluginsdk.TypeString,
												Optional:     true,
												ValidateFunc: validation.StringIsNotEmpty,
											},

											"signal_definition_name": {
												Type:         pluginsdk.TypeString,
												Optional:     true,
												ValidateFunc: validate.HealthModelChildName,
											},
										},
									},
								},
							},
						},
					},

					"azure_log_analytics": {
						Type:     pluginsdk.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &pluginsdk.Resource{
							Schema: map[string]*pluginsdk.Schema{
								"authentication_setting_name": {
									Type:         pluginsdk.TypeString,
									Required:     true,
									ValidateFunc: validate.HealthModelChildName,
								},

								"log_analytics_workspace_id": {
									Type:         pluginsdk.TypeString,
									Required:     true,
									ValidateFunc: workspaces.ValidateWorkspaceID,
								},

								"signal": {
									Type:     pluginsdk.TypeList,
									Optional: true,
									MaxItems: healthModelEntityMaxSignals,
									Elem: &pluginsdk.Resource{
										Schema: map[string]*pluginsdk.Schema{
											"name": {
												Type:         pluginsdk.TypeString,
												Required:     true,
												ValidateFunc: validate.HealthModelChildName,
											},

											"query_text": {
												Type:         pluginsdk.TypeString,
												Required:     true,
												ValidateFunc: validation.StringIsNotEmpty,
											},

											"evaluation_rule": evaluationRuleSchema(),

											"time_grain": {
												Type:         pluginsdk.TypeString,
												Optional:     true,
												ValidateFunc: validation.ISO8601Duration,
											},

											"value_column_name": {
												Type:         pluginsdk.TypeString,
												Optional:     true,
												ValidateFunc: validation.StringIsNotEmpty,
											},

											"display_name": {
												Type:         pluginsdk.TypeString,
												Optional:     true,
												ValidateFunc: validation.StringIsNotEmpty,
											},

											"refresh_interval": {
												Type:         pluginsdk.TypeString,
												Optional:     true,
												Default:      string(entities.RefreshIntervalPTOneM),
												ValidateFunc: validation.StringInSlice(validate.PossibleValuesForRefreshInterval(), false),
											},

											"data_unit": {
												Type:         pluginsdk.TypeString,
												Optional:     true,
												ValidateFunc: validation.StringIsNotEmpty,
											},

											"signal_definition_name": {
												Type:         pluginsdk.TypeString,
												Optional:     true,
												ValidateFunc: validate.HealthModelChildName,
											},
										},
									},
								},
							},
						},
					},

					"azure_monitor_workspace": {
						Type:     pluginsdk.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &pluginsdk.Resource{
							Schema: map[string]*pluginsdk.Schema{
								"authentication_setting_name": {
									Type:         pluginsdk.TypeString,
									Required:     true,
									ValidateFunc: validate.HealthModelChildName,
								},

								"monitor_workspace_id": {
									Type:         pluginsdk.TypeString,
									Required:     true,
									ValidateFunc: azuremonitorworkspaces.ValidateAccountID,
								},

								"signal": {
									Type:     pluginsdk.TypeList,
									Optional: true,
									MaxItems: healthModelEntityMaxSignals,
									Elem: &pluginsdk.Resource{
										Schema: map[string]*pluginsdk.Schema{
											"name": {
												Type:         pluginsdk.TypeString,
												Required:     true,
												ValidateFunc: validate.HealthModelChildName,
											},

											"query_text": {
												Type:         pluginsdk.TypeString,
												Required:     true,
												ValidateFunc: validation.StringIsNotEmpty,
											},

											"evaluation_rule": evaluationRuleSchema(),

											"time_grain": {
												Type:         pluginsdk.TypeString,
												Optional:     true,
												ValidateFunc: validation.ISO8601Duration,
											},

											"display_name": {
												Type:         pluginsdk.TypeString,
												Optional:     true,
												ValidateFunc: validation.StringIsNotEmpty,
											},

											"refresh_interval": {
												Type:         pluginsdk.TypeString,
												Optional:     true,
												Default:      string(entities.RefreshIntervalPTOneM),
												ValidateFunc: validation.StringInSlice(validate.PossibleValuesForRefreshInterval(), false),
											},

											"data_unit": {
												Type:         pluginsdk.TypeString,
												Optional:     true,
												ValidateFunc: validation.StringIsNotEmpty,
											},

											"signal_definition_name": {
												Type:         pluginsdk.TypeString,
												Optional:     true,
												ValidateFunc: validate.HealthModelChildName,
											},
										},
									},
								},
							},
						},
					},

					"dependency": {
						Type:     pluginsdk.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &pluginsdk.Resource{
							Schema: map[string]*pluginsdk.Schema{
								"aggregation_type": {
									Type:         pluginsdk.TypeString,
									Required:     true,
									ValidateFunc: validation.StringInSlice(validate.PossibleValuesForDependenciesAggregationType(), false),
								},

								"unit": {
									Type:         pluginsdk.TypeString,
									Optional:     true,
									ValidateFunc: validation.StringInSlice(validate.PossibleValuesForDependenciesAggregationUnit(), false),
								},

								"unhealthy_threshold": {
									Type:     pluginsdk.TypeFloat,
									Optional: true,
								},

								"degraded_threshold": {
									Type:     pluginsdk.TypeFloat,
									Optional: true,
								},

								"ignore_unknown_enabled": {
									Type:     pluginsdk.TypeBool,
									Optional: true,
									Default:  true,
								},
							},
						},
					},
				},
			},
		},

		"alert": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"unhealthy": alertConfigurationSchema(),
					"degraded":  alertConfigurationSchema(),
				},
			},
		},

		"tags": commonschema.Tags(),
	}
}

func alertConfigurationSchema() *pluginsdk.Schema {
	return &pluginsdk.Schema{
		Type:     pluginsdk.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &pluginsdk.Resource{
			Schema: map[string]*pluginsdk.Schema{
				"severity": {
					Type:         pluginsdk.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice(validate.PossibleValuesForAlertSeverity(), false),
				},

				"description": {
					Type:         pluginsdk.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringIsNotEmpty,
				},

				"action_group_ids": {
					Type:     pluginsdk.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 5,
					Elem: &pluginsdk.Schema{
						Type:         pluginsdk.TypeString,
						ValidateFunc: validation.AsGeneratedID(actiongroupsapis.ParseActionGroupIDInsensitively),
					},
				},
			},
		},
	}
}

func (HealthModelEntityResource) Attributes() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"health_state": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},

		"discovered_by": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},
	}
}

func (r HealthModelEntityResource) CustomizeDiff() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 5 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			rawConfig := metadata.ResourceDiff.GetRawConfig()

			if err := validateEvaluationRuleRawConfig(rawConfig); err != nil {
				return err
			}

			return validateDependencyRawConfig(rawConfig)
		},
	}
}

func (r HealthModelEntityResource) Create() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.CloudHealth.Entities

			var config HealthModelEntityResourceModel
			if err := metadata.Decode(&config); err != nil {
				return fmt.Errorf("decoding: %+v", err)
			}

			healthModelId, err := entities.ParseHealthmodelID(config.HealthModelId)
			if err != nil {
				return err
			}

			id := entities.NewEntityID(healthModelId.SubscriptionId, healthModelId.ResourceGroupName, healthModelId.HealthmodelName, config.Name)

			existing, err := client.Get(ctx, id)
			if err != nil && !response.WasNotFound(existing.HttpResponse) {
				return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
			}
			if !response.WasNotFound(existing.HttpResponse) {
				return metadata.ResourceRequiresImport(r.ResourceType(), id)
			}

			rawConfig, err := metadata.GetRawConfig()
			if err != nil {
				return fmt.Errorf("reading raw configuration for %s: %+v", id, err)
			}

			payload := entities.Entity{
				Properties: expandHealthModelEntityProperties(config, dependencyDegradedThresholdConfigured(rawConfig)),
			}

			if err := client.CreateOrUpdateThenPoll(ctx, id, payload); err != nil {
				return fmt.Errorf("creating %s: %+v", id, err)
			}

			metadata.SetID(id)

			return nil
		},
	}
}

func (HealthModelEntityResource) Read() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 5 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.CloudHealth.Entities

			id, err := entities.ParseEntityID(metadata.ResourceData.Id())
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

			state := HealthModelEntityResourceModel{
				Name:          id.EntityName,
				HealthModelId: entities.NewHealthmodelID(id.SubscriptionId, id.ResourceGroupName, id.HealthmodelName).ID(),
			}

			if model := resp.Model; model != nil && model.Properties != nil {
				props := *model.Properties

				state.DisplayName = pointer.From(props.DisplayName)
				state.HealthObjectivePercentage = props.HealthObjective
				state.Tags = pointer.From(props.Tags)
				state.DiscoveredBy = pointer.From(props.DiscoveredBy)

				if props.Impact != nil {
					state.Impact = string(*props.Impact)
				}

				if props.HealthState != nil {
					state.HealthState = string(*props.HealthState)
				}

				state.CanvasPosition = flattenHealthModelEntityCanvasPosition(props.CanvasPosition)
				state.Icon = flattenHealthModelEntityIcon(props.Icon)
				state.SignalGroup = flattenHealthModelEntitySignalGroups(props.SignalGroups)
				state.Alert = flattenHealthModelEntityAlerts(props.Alerts)
			}

			return metadata.Encode(&state)
		},
	}
}

func (HealthModelEntityResource) Update() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.CloudHealth.Entities

			id, err := entities.ParseEntityID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			var config HealthModelEntityResourceModel
			if err := metadata.Decode(&config); err != nil {
				return fmt.Errorf("decoding: %+v", err)
			}

			rawConfig, err := metadata.GetRawConfig()
			if err != nil {
				return fmt.Errorf("reading raw configuration for %s: %+v", *id, err)
			}

			payload := entities.Entity{
				Properties: expandHealthModelEntityProperties(config, dependencyDegradedThresholdConfigured(rawConfig)),
			}

			if err := client.CreateOrUpdateThenPoll(ctx, *id, payload); err != nil {
				return fmt.Errorf("updating %s: %+v", *id, err)
			}

			return nil
		},
	}
}

func (HealthModelEntityResource) Delete() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.CloudHealth.Entities

			id, err := entities.ParseEntityID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			resp, err := client.Delete(ctx, *id)
			if err != nil {
				// Every Health Model owns a root Entity named after the Model itself, which the service
				// creates and removes with its parent. Terraform can manage it but cannot delete it, so
				// this specific rejection means the Entity is already at its terminal state.
				if resp.OData != nil && resp.OData.Error != nil && resp.OData.Error.Code != nil && *resp.OData.Error.Code == "CannotDeleteRootEntity" {
					return nil
				}
				return fmt.Errorf("deleting %s: %+v", *id, err)
			}

			if err := resp.Poller.PollUntilDone(ctx); err != nil {
				return fmt.Errorf("polling after delete of %s: %+v", *id, err)
			}

			return nil
		},
	}
}
