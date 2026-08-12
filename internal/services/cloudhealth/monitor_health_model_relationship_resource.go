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
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/cloudhealth/sdk/2026-05-01-preview/relationships"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/cloudhealth/validate"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/validation"
)

type HealthModelRelationshipResource struct{}

var _ sdk.ResourceWithUpdate = HealthModelRelationshipResource{}

type HealthModelRelationshipResourceModel struct {
	Name             string            `tfschema:"name"`
	HealthModelId    string            `tfschema:"health_model_id"`
	ParentEntityName string            `tfschema:"parent_entity_name"`
	ChildEntityName  string            `tfschema:"child_entity_name"`
	DisplayName      string            `tfschema:"display_name"`
	Tags             map[string]string `tfschema:"tags"`
	DiscoveredBy     string            `tfschema:"discovered_by"`
}

func (HealthModelRelationshipResource) ResourceType() string {
	return "azurerm_monitor_health_model_relationship"
}

func (HealthModelRelationshipResource) ModelObject() interface{} {
	return &HealthModelRelationshipResourceModel{}
}

func (HealthModelRelationshipResource) IDValidationFunc() pluginsdk.SchemaValidateFunc {
	return relationships.ValidateRelationshipID
}

func (HealthModelRelationshipResource) Arguments() map[string]*pluginsdk.Schema {
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
			ValidateFunc: relationships.ValidateHealthModelID,
		},

		"parent_entity_name": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validate.HealthModelChildName,
		},

		"child_entity_name": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validate.HealthModelChildName,
		},

		"display_name": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringIsNotEmpty,
		},

		"tags": commonschema.Tags(),
	}
}

func (HealthModelRelationshipResource) Attributes() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"discovered_by": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},
	}
}

func (r HealthModelRelationshipResource) Create() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.CloudHealth.Relationships

			var config HealthModelRelationshipResourceModel
			if err := metadata.Decode(&config); err != nil {
				return fmt.Errorf("decoding: %+v", err)
			}

			healthModelId, err := relationships.ParseHealthModelID(config.HealthModelId)
			if err != nil {
				return err
			}

			id := relationships.NewRelationshipID(healthModelId.SubscriptionId, healthModelId.ResourceGroupName, healthModelId.HealthModelName, config.Name)

			existing, err := client.Get(ctx, id)
			if err != nil && !response.WasNotFound(existing.HttpResponse) {
				return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
			}
			if !response.WasNotFound(existing.HttpResponse) {
				return metadata.ResourceRequiresImport(r.ResourceType(), id)
			}

			payload := relationships.Relationship{
				Properties: expandHealthModelRelationshipProperties(config),
			}

			if err := client.CreateOrUpdateThenPoll(ctx, id, payload); err != nil {
				return fmt.Errorf("creating %s: %+v", id, err)
			}

			metadata.SetID(id)

			return nil
		},
	}
}

func (HealthModelRelationshipResource) Read() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 5 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.CloudHealth.Relationships

			id, err := relationships.ParseRelationshipID(metadata.ResourceData.Id())
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

			state := HealthModelRelationshipResourceModel{
				Name:          id.RelationshipName,
				HealthModelId: relationships.NewHealthModelID(id.SubscriptionId, id.ResourceGroupName, id.HealthModelName).ID(),
			}

			if model := resp.Model; model != nil && model.Properties != nil {
				props := *model.Properties

				state.ParentEntityName = props.ParentEntityName
				state.ChildEntityName = props.ChildEntityName
				state.DisplayName = pointer.From(props.DisplayName)
				state.Tags = pointer.From(props.Tags)
				state.DiscoveredBy = pointer.From(props.DiscoveredBy)
			}

			return metadata.Encode(&state)
		},
	}
}

func (HealthModelRelationshipResource) Update() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.CloudHealth.Relationships

			id, err := relationships.ParseRelationshipID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			var config HealthModelRelationshipResourceModel
			if err := metadata.Decode(&config); err != nil {
				return fmt.Errorf("decoding: %+v", err)
			}

			payload := relationships.Relationship{
				Properties: expandHealthModelRelationshipProperties(config),
			}

			if err := client.CreateOrUpdateThenPoll(ctx, *id, payload); err != nil {
				return fmt.Errorf("updating %s: %+v", *id, err)
			}

			return nil
		},
	}
}

func (HealthModelRelationshipResource) Delete() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.CloudHealth.Relationships

			id, err := relationships.ParseRelationshipID(metadata.ResourceData.Id())
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

func expandHealthModelRelationshipProperties(config HealthModelRelationshipResourceModel) *relationships.RelationshipProperties {
	props := relationships.RelationshipProperties{
		ParentEntityName: config.ParentEntityName,
		ChildEntityName:  config.ChildEntityName,
		Tags:             pointer.To(config.Tags),
	}

	if config.DisplayName != "" {
		props.DisplayName = pointer.To(config.DisplayName)
	}

	return &props
}
