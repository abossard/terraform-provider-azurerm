// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cloudhealth

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/terraform-provider-azurerm/internal/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/cloudhealth/sdk/2026-11-01/authenticationsettings"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/cloudhealth/validate"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/validation"
)

type HealthModelAuthenticationSettingResource struct{}

var _ sdk.ResourceWithUpdate = HealthModelAuthenticationSettingResource{}

type HealthModelAuthenticationSettingResourceModel struct {
	Name            string `tfschema:"name"`
	HealthModelId   string `tfschema:"health_model_id"`
	ManagedIdentity string `tfschema:"managed_identity"`
	DisplayName     string `tfschema:"display_name"`
}

func (HealthModelAuthenticationSettingResource) ResourceType() string {
	return "azurerm_monitor_health_model_authentication_setting"
}

func (HealthModelAuthenticationSettingResource) ModelObject() interface{} {
	return &HealthModelAuthenticationSettingResourceModel{}
}

func (HealthModelAuthenticationSettingResource) IDValidationFunc() pluginsdk.SchemaValidateFunc {
	return authenticationsettings.ValidateAuthenticationsettingID
}

func (HealthModelAuthenticationSettingResource) Arguments() map[string]*pluginsdk.Schema {
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
			ValidateFunc: authenticationsettings.ValidateHealthmodelID,
		},

		"managed_identity": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ValidateFunc: validation.StringIsNotEmpty,
		},

		"display_name": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringIsNotEmpty,
		},
	}
}

func (HealthModelAuthenticationSettingResource) Attributes() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{}
}

func (r HealthModelAuthenticationSettingResource) Create() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.CloudHealth.AuthenticationSettings

			var config HealthModelAuthenticationSettingResourceModel
			if err := metadata.Decode(&config); err != nil {
				return fmt.Errorf("decoding: %+v", err)
			}

			healthModelId, err := authenticationsettings.ParseHealthmodelID(config.HealthModelId)
			if err != nil {
				return err
			}

			id := authenticationsettings.NewAuthenticationsettingID(healthModelId.SubscriptionId, healthModelId.ResourceGroupName, healthModelId.HealthmodelName, config.Name)

			existing, err := client.Get(ctx, id)
			if err != nil && !response.WasNotFound(existing.HttpResponse) {
				return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
			}
			if !response.WasNotFound(existing.HttpResponse) {
				return metadata.ResourceRequiresImport(r.ResourceType(), id)
			}

			payload := authenticationsettings.AuthenticationSetting{
				Properties: expandHealthModelAuthenticationSettingProperties(config),
			}

			if err := client.CreateOrUpdateThenPoll(ctx, id, payload); err != nil {
				return fmt.Errorf("creating %s: %+v", id, err)
			}

			metadata.SetID(id)

			return nil
		},
	}
}

func (HealthModelAuthenticationSettingResource) Read() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 5 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.CloudHealth.AuthenticationSettings

			id, err := authenticationsettings.ParseAuthenticationsettingID(metadata.ResourceData.Id())
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

			state := HealthModelAuthenticationSettingResourceModel{
				Name:          id.AuthenticationsettingName,
				HealthModelId: authenticationsettings.NewHealthmodelID(id.SubscriptionId, id.ResourceGroupName, id.HealthmodelName).ID(),
			}

			if model := resp.Model; model != nil && model.Properties != nil {
				props, ok := model.Properties.(authenticationsettings.ManagedIdentityAuthenticationSettingProperties)
				if !ok {
					return fmt.Errorf("retrieving %s: expected properties of kind %q but got %q", *id, authenticationsettings.AuthenticationKindManagedIdentity, model.Properties.AuthenticationSettingProperties().AuthenticationKind)
				}

				state.ManagedIdentity = props.ManagedIdentityName
				state.DisplayName = pointer.From(props.DisplayName)
			}

			return metadata.Encode(&state)
		},
	}
}

func (HealthModelAuthenticationSettingResource) Update() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.CloudHealth.AuthenticationSettings

			id, err := authenticationsettings.ParseAuthenticationsettingID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			var config HealthModelAuthenticationSettingResourceModel
			if err := metadata.Decode(&config); err != nil {
				return fmt.Errorf("decoding: %+v", err)
			}

			payload := authenticationsettings.AuthenticationSetting{
				Properties: expandHealthModelAuthenticationSettingProperties(config),
			}

			if err := client.CreateOrUpdateThenPoll(ctx, *id, payload); err != nil {
				return fmt.Errorf("updating %s: %+v", *id, err)
			}

			return nil
		},
	}
}

func (HealthModelAuthenticationSettingResource) Delete() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.CloudHealth.AuthenticationSettings

			id, err := authenticationsettings.ParseAuthenticationsettingID(metadata.ResourceData.Id())
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

func expandHealthModelAuthenticationSettingProperties(config HealthModelAuthenticationSettingResourceModel) authenticationsettings.AuthenticationSettingProperties {
	props := authenticationsettings.ManagedIdentityAuthenticationSettingProperties{
		AuthenticationKind:  authenticationsettings.AuthenticationKindManagedIdentity,
		ManagedIdentityName: config.ManagedIdentity,
	}

	if config.DisplayName != "" {
		props.DisplayName = pointer.To(config.DisplayName)
	}

	return props
}
