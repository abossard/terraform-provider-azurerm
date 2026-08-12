// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cloudhealth

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	components "github.com/hashicorp/go-azure-sdk/resource-manager/applicationinsights/2020-02-02/componentsapis"
	"github.com/hashicorp/terraform-provider-azurerm/internal/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/cloudhealth/sdk/2026-05-01-preview/discoveryrules"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/cloudhealth/validate"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/validation"
)

type HealthModelDiscoveryRuleResource struct{}

var _ sdk.ResourceWithUpdate = HealthModelDiscoveryRuleResource{}

type HealthModelDiscoveryRuleResourceModel struct {
	Name                         string `tfschema:"name"`
	HealthModelId                string `tfschema:"health_model_id"`
	AuthenticationSettingName    string `tfschema:"authentication_setting_name"`
	ResourceGraphQuery           string `tfschema:"resource_graph_query"`
	ApplicationInsightsId        string `tfschema:"application_insights_id"`
	DisplayName                  string `tfschema:"display_name"`
	RelationshipDiscoveryEnabled bool   `tfschema:"relationship_discovery_enabled"`
	RecommendedSignalsEnabled    bool   `tfschema:"recommended_signals_enabled"`
	ResourceHealthSignalEnabled  bool   `tfschema:"resource_health_signal_enabled"`
	EntityName                   string `tfschema:"entity_name"`
}

func (HealthModelDiscoveryRuleResource) ResourceType() string {
	return "azurerm_monitor_health_model_discovery_rule"
}

func (HealthModelDiscoveryRuleResource) ModelObject() interface{} {
	return &HealthModelDiscoveryRuleResourceModel{}
}

func (HealthModelDiscoveryRuleResource) IDValidationFunc() pluginsdk.SchemaValidateFunc {
	return discoveryrules.ValidateDiscoveryRuleID
}

func (HealthModelDiscoveryRuleResource) Arguments() map[string]*pluginsdk.Schema {
	specifications := []string{"resource_graph_query", "application_insights_id"}

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
			ValidateFunc: discoveryrules.ValidateHealthModelID,
		},

		"authentication_setting_name": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ValidateFunc: validate.HealthModelChildName,
		},

		"resource_graph_query": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			ForceNew:     true,
			ExactlyOneOf: specifications,
			ValidateFunc: validation.StringIsNotEmpty,
		},

		"application_insights_id": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			ForceNew:     true,
			ExactlyOneOf: specifications,
			ValidateFunc: components.ValidateComponentID,
		},

		"display_name": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringIsNotEmpty,
		},

		"relationship_discovery_enabled": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
			Default:  true,
		},

		"recommended_signals_enabled": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
			Default:  true,
		},

		"resource_health_signal_enabled": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
			Default:  true,
		},
	}
}

func (HealthModelDiscoveryRuleResource) Attributes() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"entity_name": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},
	}
}

func (r HealthModelDiscoveryRuleResource) Create() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.CloudHealth.DiscoveryRules

			var config HealthModelDiscoveryRuleResourceModel
			if err := metadata.Decode(&config); err != nil {
				return fmt.Errorf("decoding: %+v", err)
			}

			healthModelId, err := discoveryrules.ParseHealthModelID(config.HealthModelId)
			if err != nil {
				return err
			}

			id := discoveryrules.NewDiscoveryRuleID(healthModelId.SubscriptionId, healthModelId.ResourceGroupName, healthModelId.HealthModelName, config.Name)

			existing, err := client.Get(ctx, id)
			if err != nil && !response.WasNotFound(existing.HttpResponse) {
				return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
			}
			if !response.WasNotFound(existing.HttpResponse) {
				return metadata.ResourceRequiresImport(r.ResourceType(), id)
			}

			props, err := expandHealthModelDiscoveryRuleProperties(config)
			if err != nil {
				return err
			}

			payload := discoveryrules.DiscoveryRule{
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

func (HealthModelDiscoveryRuleResource) Read() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 5 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.CloudHealth.DiscoveryRules

			id, err := discoveryrules.ParseDiscoveryRuleID(metadata.ResourceData.Id())
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

			state := HealthModelDiscoveryRuleResourceModel{
				Name:          id.DiscoveryRuleName,
				HealthModelId: discoveryrules.NewHealthModelID(id.SubscriptionId, id.ResourceGroupName, id.HealthModelName).ID(),
			}

			if model := resp.Model; model != nil && model.Properties != nil {
				props := *model.Properties

				state.AuthenticationSettingName = props.AuthenticationSetting
				state.EntityName = props.EntityName
				state.DisplayName = pointer.From(props.DisplayName)
				state.RelationshipDiscoveryEnabled = props.DiscoverRelationships == discoveryrules.DiscoveryRuleRelationshipDiscoveryBehaviorEnabled
				state.RecommendedSignalsEnabled = props.AddRecommendedSignals == discoveryrules.DiscoveryRuleRecommendedSignalsBehaviorEnabled

				// The service omits `addResourceHealthSignal` when it holds its default value, so an
				// absent field reads back as enabled rather than as false.
				state.ResourceHealthSignalEnabled = true
				if props.AddResourceHealthSignal != nil {
					state.ResourceHealthSignalEnabled = *props.AddResourceHealthSignal == discoveryrules.ResourceHealthAvailabilityStateSignalBehaviorEnabled
				}

				switch specification := props.Specification.(type) {
				case discoveryrules.ResourceGraphQuerySpecification:
					state.ResourceGraphQuery = specification.ResourceGraphQuery
				case discoveryrules.ApplicationInsightsTopologySpecification:
					state.ApplicationInsightsId = specification.ApplicationInsightsResourceId
				default:
					return fmt.Errorf("retrieving %s: unhandled specification kind %q", *id, props.Specification.DiscoveryRuleSpecification().Kind)
				}
			}

			return metadata.Encode(&state)
		},
	}
}

func (HealthModelDiscoveryRuleResource) Update() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.CloudHealth.DiscoveryRules

			id, err := discoveryrules.ParseDiscoveryRuleID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			var config HealthModelDiscoveryRuleResourceModel
			if err := metadata.Decode(&config); err != nil {
				return fmt.Errorf("decoding: %+v", err)
			}

			props, err := expandHealthModelDiscoveryRuleProperties(config)
			if err != nil {
				return err
			}

			payload := discoveryrules.DiscoveryRule{
				Properties: props,
			}

			if err := client.CreateOrUpdateThenPoll(ctx, *id, payload); err != nil {
				return fmt.Errorf("updating %s: %+v", *id, err)
			}

			return nil
		},
	}
}

func (HealthModelDiscoveryRuleResource) Delete() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.CloudHealth.DiscoveryRules

			id, err := discoveryrules.ParseDiscoveryRuleID(metadata.ResourceData.Id())
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

func expandHealthModelDiscoveryRuleProperties(config HealthModelDiscoveryRuleResourceModel) (*discoveryrules.DiscoveryRuleProperties, error) {
	discoverRelationships := discoveryrules.DiscoveryRuleRelationshipDiscoveryBehaviorDisabled
	if config.RelationshipDiscoveryEnabled {
		discoverRelationships = discoveryrules.DiscoveryRuleRelationshipDiscoveryBehaviorEnabled
	}

	addRecommendedSignals := discoveryrules.DiscoveryRuleRecommendedSignalsBehaviorDisabled
	if config.RecommendedSignalsEnabled {
		addRecommendedSignals = discoveryrules.DiscoveryRuleRecommendedSignalsBehaviorEnabled
	}

	addResourceHealthSignal := discoveryrules.ResourceHealthAvailabilityStateSignalBehaviorDisabled
	if config.ResourceHealthSignalEnabled {
		addResourceHealthSignal = discoveryrules.ResourceHealthAvailabilityStateSignalBehaviorEnabled
	}

	props := discoveryrules.DiscoveryRuleProperties{
		AuthenticationSetting:   config.AuthenticationSettingName,
		DiscoverRelationships:   discoverRelationships,
		AddRecommendedSignals:   addRecommendedSignals,
		AddResourceHealthSignal: pointer.To(addResourceHealthSignal),
	}

	if config.DisplayName != "" {
		props.DisplayName = pointer.To(config.DisplayName)
	}

	switch {
	case config.ResourceGraphQuery != "":
		props.Specification = discoveryrules.ResourceGraphQuerySpecification{
			Kind:               discoveryrules.DiscoveryRuleKindResourceGraphQuery,
			ResourceGraphQuery: config.ResourceGraphQuery,
		}
	case config.ApplicationInsightsId != "":
		props.Specification = discoveryrules.ApplicationInsightsTopologySpecification{
			Kind:                          discoveryrules.DiscoveryRuleKindApplicationInsightsTopology,
			ApplicationInsightsResourceId: config.ApplicationInsightsId,
		}
	default:
		return nil, fmt.Errorf("exactly one of `resource_graph_query` or `application_insights_id` must be set")
	}

	return &props, nil
}
