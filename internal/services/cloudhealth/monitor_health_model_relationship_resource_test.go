// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cloudhealth_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/internal/clients"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/cloudhealth/parse"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
)

type MonitorHealthModelRelationshipResource struct{}

func TestAccMonitorHealthModelRelationship_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model_relationship", "test")
	r := MonitorHealthModelRelationshipResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccMonitorHealthModelRelationship_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model_relationship", "test")
	r := MonitorHealthModelRelationshipResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.RequiresImportErrorStep(r.requiresImport),
	})
}

func TestAccMonitorHealthModelRelationship_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model_relationship", "test")
	r := MonitorHealthModelRelationshipResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

// TestAccMonitorHealthModelRelationship_update covers the only updatable property on this resource.
// `parentEntityName` and `childEntityName` are declared `x-ms-mutability: ["read", "create"]` in the
// 2026-05-01-preview swagger, so changing either must force a new resource rather than update in place.
func TestAccMonitorHealthModelRelationship_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model_relationship", "test")
	r := MonitorHealthModelRelationshipResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (r MonitorHealthModelRelationshipResource) Exists(_ context.Context, _ *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.RelationshipID(state.ID)
	if err != nil {
		return nil, err
	}

	// TODO: replace with `clients.CloudHealth.RelationshipsClient.Get(ctx, *id)` once a stable
	// `Microsoft.CloudHealth` API version ships and `hashicorp/go-azure-sdk` is vendored.
	return nil, fmt.Errorf("checking for %s: %w", id, errSDKUnavailable)
}

func (r MonitorHealthModelRelationshipResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctest-rg-%[1]d"
  location = "%[2]s"
}

resource "azurerm_storage_account" "test" {
  name                     = "acctestsa%[3]s"
  resource_group_name      = azurerm_resource_group.test.name
  location                 = azurerm_resource_group.test.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurerm_monitor_health_model" "test" {
  name                = "acctest-hm-%[1]d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location

  identity {
    type = "SystemAssigned"
  }
}

resource "azurerm_monitor_health_model_authentication_setting" "test" {
  name             = "acctest-as-%[1]d"
  health_model_id  = azurerm_monitor_health_model.test.id
  managed_identity = "SystemAssigned"
}

# The CloudHealth resource provider creates a root entity named exactly like the health model, so the
# parent entity below manages that built-in root rather than creating a second, competing root.
resource "azurerm_monitor_health_model_entity" "parent" {
  name            = azurerm_monitor_health_model.test.name
  health_model_id = azurerm_monitor_health_model.test.id
  display_name    = "Workload"

  signal_group {
    dependency {
      aggregation_type = "WorstOf"
    }
  }
}

resource "azurerm_monitor_health_model_entity" "child" {
  name            = "acctest-entity-child-%[1]d"
  health_model_id = azurerm_monitor_health_model.test.id
  display_name    = "Storage Account"

  signal_group {
    azure_resource {
      authentication_setting_name = azurerm_monitor_health_model_authentication_setting.test.name
      target_resource_id          = azurerm_storage_account.test.id
      resource_health_enabled     = true
    }
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString)
}

func (r MonitorHealthModelRelationshipResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_monitor_health_model_relationship" "test" {
  name               = "acctest-rel-%d"
  health_model_id    = azurerm_monitor_health_model.test.id
  parent_entity_name = azurerm_monitor_health_model_entity.parent.name
  child_entity_name  = azurerm_monitor_health_model_entity.child.name
}
`, r.template(data), data.RandomInteger)
}

func (r MonitorHealthModelRelationshipResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_monitor_health_model_relationship" "import" {
  name               = azurerm_monitor_health_model_relationship.test.name
  health_model_id    = azurerm_monitor_health_model_relationship.test.health_model_id
  parent_entity_name = azurerm_monitor_health_model_relationship.test.parent_entity_name
  child_entity_name  = azurerm_monitor_health_model_relationship.test.child_entity_name
}
`, r.basic(data))
}

func (r MonitorHealthModelRelationshipResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_monitor_health_model_relationship" "test" {
  name               = "acctest-rel-%d"
  health_model_id    = azurerm_monitor_health_model.test.id
  parent_entity_name = azurerm_monitor_health_model_entity.parent.name
  child_entity_name  = azurerm_monitor_health_model_entity.child.name
  display_name       = "Workload depends on Storage Account"

  tags = {
    tier = "data"
  }
}
`, r.template(data), data.RandomInteger)
}
