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
	"github.com/hashicorp/go-azure-helpers/resourcemanager/identity"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-provider-azurerm/internal/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/cloudhealth/sdk/2026-11-01/healthmodels"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/cloudhealth/validate"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
)

type HealthModelResource struct{}

var _ sdk.ResourceWithUpdate = HealthModelResource{}

type HealthModelResourceModel struct {
	Name              string                                     `tfschema:"name"`
	ResourceGroupName string                                     `tfschema:"resource_group_name"`
	Location          string                                     `tfschema:"location"`
	Identity          []identity.ModelSystemAssignedUserAssigned `tfschema:"identity"`
	Tags              map[string]string                          `tfschema:"tags"`
}

func (HealthModelResource) ResourceType() string {
	return "azurerm_monitor_health_model"
}

func (HealthModelResource) ModelObject() interface{} {
	return &HealthModelResourceModel{}
}

func (HealthModelResource) IDValidationFunc() pluginsdk.SchemaValidateFunc {
	return healthmodels.ValidateHealthmodelID
}

func (HealthModelResource) Arguments() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validate.HealthModelName,
		},

		"resource_group_name": commonschema.ResourceGroupName(),

		"location": commonschema.Location(),

		"identity": commonschema.SystemAssignedUserAssignedIdentityOptional(),

		"tags": commonschema.Tags(),
	}
}

func (HealthModelResource) Attributes() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{}
}

func (r HealthModelResource) Create() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.CloudHealth.HealthModels
			subscriptionId := metadata.Client.Account.SubscriptionId

			var config HealthModelResourceModel
			if err := metadata.Decode(&config); err != nil {
				return fmt.Errorf("decoding: %+v", err)
			}

			id := healthmodels.NewHealthmodelID(subscriptionId, config.ResourceGroupName, config.Name)

			existing, err := client.Get(ctx, id)
			if err != nil && !response.WasNotFound(existing.HttpResponse) {
				return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
			}
			if !response.WasNotFound(existing.HttpResponse) {
				return metadata.ResourceRequiresImport(r.ResourceType(), id)
			}

			expandedIdentity, err := identity.ExpandLegacySystemAndUserAssignedMapFromModel(config.Identity)
			if err != nil {
				return fmt.Errorf("expanding `identity`: %+v", err)
			}

			payload := healthmodels.HealthModel{
				Location:   location.Normalize(config.Location),
				Identity:   expandedIdentity,
				Properties: &healthmodels.HealthModelProperties{},
				Tags:       pointer.To(config.Tags),
			}

			if err := client.CreateThenPoll(ctx, id, payload); err != nil {
				return fmt.Errorf("creating %s: %+v", id, err)
			}

			metadata.SetID(id)

			return nil
		},
	}
}

func (HealthModelResource) Read() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 5 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.CloudHealth.HealthModels

			id, err := healthmodels.ParseHealthmodelID(metadata.ResourceData.Id())
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

			state := HealthModelResourceModel{
				Name:              id.HealthmodelName,
				ResourceGroupName: id.ResourceGroupName,
			}

			if model := resp.Model; model != nil {
				state.Location = location.Normalize(model.Location)
				state.Tags = pointer.From(model.Tags)

				flattenedIdentity, err := identity.FlattenLegacySystemAndUserAssignedMapToModel(model.Identity)
				if err != nil {
					return fmt.Errorf("flattening `identity`: %+v", err)
				}
				state.Identity = flattenedIdentity
			}

			return metadata.Encode(&state)
		},
	}
}

func (HealthModelResource) Update() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.CloudHealth.HealthModels

			id, err := healthmodels.ParseHealthmodelID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			var config HealthModelResourceModel
			if err := metadata.Decode(&config); err != nil {
				return fmt.Errorf("decoding: %+v", err)
			}

			payload := healthmodels.HealthModelUpdate{}

			if metadata.ResourceData.HasChange("identity") {
				expandedIdentity, err := identity.ExpandLegacySystemAndUserAssignedMapFromModel(config.Identity)
				if err != nil {
					return fmt.Errorf("expanding `identity`: %+v", err)
				}
				payload.Identity = expandedIdentity
			}

			if metadata.ResourceData.HasChange("tags") {
				payload.Tags = pointer.To(config.Tags)
			}

			if err := client.UpdateThenPoll(ctx, *id, payload); err != nil {
				return fmt.Errorf("updating %s: %+v", *id, err)
			}

			return nil
		},
	}
}

func (HealthModelResource) Delete() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.CloudHealth.HealthModels

			id, err := healthmodels.ParseHealthmodelID(metadata.ResourceData.Id())
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
