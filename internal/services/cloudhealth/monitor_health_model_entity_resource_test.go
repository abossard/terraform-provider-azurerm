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

type MonitorHealthModelEntityResource struct{}

func TestAccMonitorHealthModelEntity_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model_entity", "test")
	r := MonitorHealthModelEntityResource{}
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

func TestAccMonitorHealthModelEntity_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model_entity", "test")
	r := MonitorHealthModelEntityResource{}
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

func TestAccMonitorHealthModelEntity_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model_entity", "test")
	r := MonitorHealthModelEntityResource{}
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

func TestAccMonitorHealthModelEntity_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model_entity", "test")
	r := MonitorHealthModelEntityResource{}
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

// TestAccMonitorHealthModelEntity_dependencyAggregation covers the grouping-entity shape, where health
// rolls up from child entities instead of from an Azure resource's own signals.
func TestAccMonitorHealthModelEntity_dependencyAggregation(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model_entity", "test")
	r := MonitorHealthModelEntityResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.dependencyAggregation(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (r MonitorHealthModelEntityResource) Exists(_ context.Context, _ *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.EntityID(state.ID)
	if err != nil {
		return nil, err
	}

	// TODO: replace with `clients.CloudHealth.EntitiesClient.Get(ctx, *id)` once a stable
	// `Microsoft.CloudHealth` API version ships and `hashicorp/go-azure-sdk` is vendored.
	return nil, fmt.Errorf("checking for %s: %w", id, errSDKUnavailable)
}

func (r MonitorHealthModelEntityResource) template(data acceptance.TestData) string {
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
`, data.RandomInteger, data.Locations.Primary, data.RandomString)
}

func (r MonitorHealthModelEntityResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_monitor_health_model_entity" "test" {
  name            = "acctest-entity-%d"
  health_model_id = azurerm_monitor_health_model.test.id
}
`, r.template(data), data.RandomInteger)
}

func (r MonitorHealthModelEntityResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_monitor_health_model_entity" "import" {
  name            = azurerm_monitor_health_model_entity.test.name
  health_model_id = azurerm_monitor_health_model_entity.test.health_model_id
}
`, r.basic(data))
}

func (r MonitorHealthModelEntityResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_monitor_action_group" "test" {
  name                = "acctest-ag-%[2]d"
  resource_group_name = azurerm_resource_group.test.name
  short_name          = "acctestag"
}

resource "azurerm_monitor_health_model_entity" "test" {
  name                        = "acctest-entity-%[2]d"
  health_model_id             = azurerm_monitor_health_model.test.id
  display_name                = "Storage Account"
  health_objective_percentage = 99.9
  impact                      = "Standard"

  canvas_position {
    x = 120
    y = 480
  }

  icon {
    name = "StorageAccount"
  }

  signal_group {
    azure_resource {
      authentication_setting_name = azurerm_monitor_health_model_authentication_setting.test.name
      target_resource_id          = azurerm_storage_account.test.id
      resource_health_enabled     = true

      signal {
        name             = "availability"
        display_name     = "Availability"
        metric_namespace = "Microsoft.Storage/storageAccounts"
        metric_name      = "Availability"
        aggregation_type = "Average"
        time_grain       = "PT5M"
        refresh_interval = "PT5M"
        data_unit        = "Percent"

        evaluation_rule {
          degraded {
            operator  = "LessThan"
            threshold = 99.9
          }

          unhealthy {
            operator  = "LessThan"
            threshold = 99
          }
        }
      }

      signal {
        name             = "throttled-requests"
        metric_namespace = "Microsoft.Storage/storageAccounts"
        metric_name      = "Transactions"
        aggregation_type = "Total"
        time_grain       = "PT5M"
        refresh_interval = "PT5M"
        dimension_filter = "ResponseType eq 'ServerBusyError'"
        data_unit        = "Count"

        evaluation_rule {
          unhealthy {
            operator         = "Dynamic"
            sensitivity      = "Medium"
            look_back_window = "PT1H"
          }
        }
      }
    }
  }

  alert {
    unhealthy {
      severity         = "Sev1"
      description      = "The storage account is unhealthy"
      action_group_ids = [azurerm_monitor_action_group.test.id]
    }

    degraded {
      severity         = "Sev3"
      action_group_ids = [azurerm_monitor_action_group.test.id]
    }
  }

  tags = {
    tier = "data"
  }
}
`, r.template(data), data.RandomInteger)
}

func (r MonitorHealthModelEntityResource) dependencyAggregation(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_monitor_health_model_entity" "test" {
  name            = "acctest-entity-%[2]d"
  health_model_id = azurerm_monitor_health_model.test.id
  display_name    = "Data Tier"

  canvas_position {
    x = 0
    y = 240
  }

  signal_group {
    dependency {
      aggregation_type       = "MinHealthy"
      unit                   = "Percentage"
      degraded_threshold     = 80
      unhealthy_threshold    = 50
      ignore_unknown_enabled = true
    }
  }
}
`, r.template(data), data.RandomInteger)
}
