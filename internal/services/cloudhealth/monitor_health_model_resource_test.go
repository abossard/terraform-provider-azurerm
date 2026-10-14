// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cloudhealth_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/internal/clients"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/cloudhealth/sdk/2026-11-01/healthmodels"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
)

type MonitorHealthModelResource struct{}

func TestAccMonitorHealthModel_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model", "test")
	r := MonitorHealthModelResource{}
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

func TestAccMonitorHealthModel_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model", "test")
	r := MonitorHealthModelResource{}
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

func TestAccMonitorHealthModel_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model", "test")
	r := MonitorHealthModelResource{}
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

func TestAccMonitorHealthModel_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model", "test")
	r := MonitorHealthModelResource{}
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
			Config: r.update(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (r MonitorHealthModelResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := healthmodels.ParseHealthmodelID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.CloudHealth.HealthModels.Get(ctx, *id)
	if err != nil {
		if response.WasNotFound(resp.HttpResponse) {
			return pointer.To(false), nil
		}
		return nil, fmt.Errorf("retrieving %s: %+v", *id, err)
	}

	return pointer.To(resp.Model != nil), nil
}

func (r MonitorHealthModelResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctest-rg-%d"
  location = "%s"
}
`, data.RandomInteger, data.Locations.Primary)
}

func (r MonitorHealthModelResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_monitor_health_model" "test" {
  name                = "acctest-hm-%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
}
`, r.template(data), data.RandomInteger)
}

func (r MonitorHealthModelResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_monitor_health_model" "import" {
  name                = azurerm_monitor_health_model.test.name
  resource_group_name = azurerm_monitor_health_model.test.resource_group_name
  location            = azurerm_monitor_health_model.test.location
}
`, r.basic(data))
}

func (r MonitorHealthModelResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_user_assigned_identity" "test" {
  name                = "acctest-uai-%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
}

resource "azurerm_monitor_health_model" "test" {
  name                = "acctest-hm-%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location

  identity {
    type         = "UserAssigned"
    identity_ids = [azurerm_user_assigned_identity.test.id]
  }

  tags = {
    environment = "acctest"
  }
}
`, r.template(data), data.RandomInteger, data.RandomInteger)
}

func (r MonitorHealthModelResource) update(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_monitor_health_model" "test" {
  name                = "acctest-hm-%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location

  identity {
    type = "SystemAssigned"
  }

  tags = {
    environment = "acctest-updated"
    cost_centre = "1234"
  }
}
`, r.template(data), data.RandomInteger)
}
