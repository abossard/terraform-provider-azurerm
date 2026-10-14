// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"fmt"

	"github.com/hashicorp/terraform-provider-azurerm/internal/common"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/cloudhealth/sdk/2026-11-01/authenticationsettings"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/cloudhealth/sdk/2026-11-01/discoveryrules"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/cloudhealth/sdk/2026-11-01/entities"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/cloudhealth/sdk/2026-11-01/healthmodels"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/cloudhealth/sdk/2026-11-01/relationships"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/cloudhealth/sdk/2026-11-01/signaldefinitions"
)

// Client holds the SDK clients for the `Microsoft.CloudHealth` service.
//
// The clients use the stable 2026-11-01 API from the local Pandora-generated SDK until CloudHealth
// is available in the vendored `hashicorp/go-azure-sdk`.
type Client struct {
	AuthenticationSettings *authenticationsettings.AuthenticationSettingsClient
	DiscoveryRules         *discoveryrules.DiscoveryRulesClient
	Entities               *entities.EntitiesClient
	HealthModels           *healthmodels.HealthModelsClient
	Relationships          *relationships.RelationshipsClient
	SignalDefinitions      *signaldefinitions.SignalDefinitionsClient
}

func NewClient(o *common.ClientOptions) (*Client, error) {
	authenticationSettingsClient, err := authenticationsettings.NewAuthenticationSettingsClientWithBaseURI(o.Environment.ResourceManager)
	if err != nil {
		return nil, fmt.Errorf("building AuthenticationSettings client: %+v", err)
	}
	o.Configure(authenticationSettingsClient.Client, o.Authorizers.ResourceManager)

	discoveryRulesClient, err := discoveryrules.NewDiscoveryRulesClientWithBaseURI(o.Environment.ResourceManager)
	if err != nil {
		return nil, fmt.Errorf("building DiscoveryRules client: %+v", err)
	}
	o.Configure(discoveryRulesClient.Client, o.Authorizers.ResourceManager)

	entitiesClient, err := entities.NewEntitiesClientWithBaseURI(o.Environment.ResourceManager)
	if err != nil {
		return nil, fmt.Errorf("building Entities client: %+v", err)
	}
	o.Configure(entitiesClient.Client, o.Authorizers.ResourceManager)

	healthModelsClient, err := healthmodels.NewHealthModelsClientWithBaseURI(o.Environment.ResourceManager)
	if err != nil {
		return nil, fmt.Errorf("building HealthModels client: %+v", err)
	}
	o.Configure(healthModelsClient.Client, o.Authorizers.ResourceManager)

	relationshipsClient, err := relationships.NewRelationshipsClientWithBaseURI(o.Environment.ResourceManager)
	if err != nil {
		return nil, fmt.Errorf("building Relationships client: %+v", err)
	}
	o.Configure(relationshipsClient.Client, o.Authorizers.ResourceManager)

	signalDefinitionsClient, err := signaldefinitions.NewSignalDefinitionsClientWithBaseURI(o.Environment.ResourceManager)
	if err != nil {
		return nil, fmt.Errorf("building SignalDefinitions client: %+v", err)
	}
	o.Configure(signalDefinitionsClient.Client, o.Authorizers.ResourceManager)

	return &Client{
		AuthenticationSettings: authenticationSettingsClient,
		DiscoveryRules:         discoveryRulesClient,
		Entities:               entitiesClient,
		HealthModels:           healthModelsClient,
		Relationships:          relationshipsClient,
		SignalDefinitions:      signalDefinitionsClient,
	}, nil
}
