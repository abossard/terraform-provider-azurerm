// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cloudhealth_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/internal/clients"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/cloudhealth/sdk/2026-11-01/entities"
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

func TestAccMonitorHealthModelEntity_zeroValues(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model_entity", "test")
	r := MonitorHealthModelEntityResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.zeroValues(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("health_objective_percentage").HasValue("0"),
				check.That(data.ResourceName).Key("signal_group.0.dependency.0.degraded_threshold").HasValue("0"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccMonitorHealthModelEntity_signalCoverage(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model_entity", "test")
	r := MonitorHealthModelEntityResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.signalCoverage(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccMonitorHealthModelEntity_dynamicOperatorWithThreshold(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model_entity", "test")
	r := MonitorHealthModelEntityResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config:      r.dynamicOperatorWithThreshold(data),
			PlanOnly:    true,
			ExpectError: regexp.MustCompile("`threshold` cannot be set in `evaluation_rule.unhealthy` when `operator` is `Dynamic`"),
		},
	})
}

func TestAccMonitorHealthModelEntity_worstOfWithThreshold(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model_entity", "test")
	r := MonitorHealthModelEntityResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config:      r.worstOfWithThreshold(data),
			PlanOnly:    true,
			ExpectError: regexp.MustCompile("`unhealthy_threshold` cannot be set in `signal_group.dependency` when `aggregation_type` is `WorstOf`"),
		},
	})
}

func TestAccMonitorHealthModelEntity_countingAggregationWithoutThreshold(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model_entity", "test")
	r := MonitorHealthModelEntityResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config:      r.countingAggregationWithoutThreshold(data),
			PlanOnly:    true,
			ExpectError: regexp.MustCompile("`unhealthy_threshold` must be set in `signal_group.dependency` when `aggregation_type` is `MinHealthy`"),
		},
	})
}

func TestAccMonitorHealthModelEntity_countingAggregationWithoutUnit(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model_entity", "test")
	r := MonitorHealthModelEntityResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config:      r.countingAggregationWithoutUnit(data),
			PlanOnly:    true,
			ExpectError: regexp.MustCompile("`unit` must be set in `signal_group.dependency` when `aggregation_type` is `MaxNotHealthy`"),
		},
	})
}

func (r MonitorHealthModelEntityResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := entities.ParseEntityID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.CloudHealth.Entities.Get(ctx, *id)
	if err != nil {
		if response.WasNotFound(resp.HttpResponse) {
			return pointer.To(false), nil
		}
		return nil, fmt.Errorf("retrieving %s: %+v", *id, err)
	}

	return pointer.To(resp.Model != nil), nil
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

  public_network_access_enabled = false
  shared_access_key_enabled     = false
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

// zeroValues pins the zero-versus-absent distinction end to end. `0` is a legal, documented value
// for both `health_objective_percentage` (`FloatBetween(0, 100)`) and `degraded_threshold`, and the
// service stores both, so a post-apply plan must be empty rather than showing the value being
// re-sent. `MaxNotHealthy` is used because it is the aggregation whose ordering rule allows a
// degraded threshold below the unhealthy one.
func (r MonitorHealthModelEntityResource) zeroValues(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_monitor_health_model_entity" "test" {
  name            = "acctest-entity-%d"
  health_model_id = azurerm_monitor_health_model.test.id

  health_objective_percentage = 0

  signal_group {
    dependency {
      aggregation_type    = "MaxNotHealthy"
      unit                = "Absolute"
      unhealthy_threshold = 1
      degraded_threshold  = 0
    }
  }
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
            operator    = "Dynamic"
            sensitivity = "Medium"
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

func (r MonitorHealthModelEntityResource) worstOfWithThreshold(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_monitor_health_model_entity" "test" {
  name            = "acctest-entity-%[2]d"
  health_model_id = azurerm_monitor_health_model.test.id

  signal_group {
    dependency {
      aggregation_type    = "WorstOf"
      unhealthy_threshold = 50
    }
  }
}
`, r.template(data), data.RandomInteger)
}

func (r MonitorHealthModelEntityResource) countingAggregationWithoutThreshold(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_monitor_health_model_entity" "test" {
  name            = "acctest-entity-%[2]d"
  health_model_id = azurerm_monitor_health_model.test.id

  signal_group {
    dependency {
      aggregation_type = "MinHealthy"
      unit             = "Percentage"
    }
  }
}
`, r.template(data), data.RandomInteger)
}

func (r MonitorHealthModelEntityResource) countingAggregationWithoutUnit(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_monitor_health_model_entity" "test" {
  name            = "acctest-entity-%[2]d"
  health_model_id = azurerm_monitor_health_model.test.id

  signal_group {
    dependency {
      aggregation_type    = "MaxNotHealthy"
      unhealthy_threshold = 25
    }
  }
}
`, r.template(data), data.RandomInteger)
}

func (r MonitorHealthModelEntityResource) signalCoverage(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_log_analytics_workspace" "test" {
  name                = "acctest-law-%[2]d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  sku                 = "PerGB2018"
  retention_in_days   = 30
}

resource "azurerm_monitor_workspace" "test" {
  name                = "acctest-amw-%[2]d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
}

resource "azurerm_monitor_health_model_signal_definition" "metric" {
  name            = "acctest-sd-metric-%[2]d"
  health_model_id = azurerm_monitor_health_model.test.id

  azure_resource_metric {
    metric_namespace = "Microsoft.Storage/storageAccounts"
    metric_name      = "Availability"
    aggregation_type = "Average"
    time_grain       = "PT5M"
  }

  evaluation_rule {
    unhealthy {
      operator  = "LessThan"
      threshold = 99
    }
  }
}

resource "azurerm_monitor_health_model_signal_definition" "log" {
  name            = "acctest-sd-log-%[2]d"
  health_model_id = azurerm_monitor_health_model.test.id

  log_analytics_query {
    query_text        = "Heartbeat | summarize Total = count() by bin(TimeGenerated, 5m)"
    time_grain        = "PT5M"
    value_column_name = "Total"
  }

  evaluation_rule {
    unhealthy {
      operator  = "LessThan"
      threshold = 1
    }
  }
}

resource "azurerm_monitor_health_model_signal_definition" "prometheus" {
  name            = "acctest-sd-prom-%[2]d"
  health_model_id = azurerm_monitor_health_model.test.id

  prometheus_metrics_query {
    query_text = "up"
    time_grain = "PT5M"
  }

  evaluation_rule {
    unhealthy {
      operator  = "LessThan"
      threshold = 1
    }
  }
}

resource "azurerm_monitor_health_model_entity" "test" {
  name            = "acctest-entity-%[2]d"
  health_model_id = azurerm_monitor_health_model.test.id
  display_name    = "Signal coverage"

  icon {
    name               = "Custom"
    custom_data_base64 = "%[3]s"
  }

  signal_group {
    azure_resource {
      authentication_setting_name = azurerm_monitor_health_model_authentication_setting.test.name
      target_resource_id          = azurerm_storage_account.test.id

      signal {
        name                   = "availability"
        metric_namespace       = "Microsoft.Storage/storageAccounts"
        metric_name            = "Availability"
        aggregation_type       = "Average"
        time_grain             = "PT5M"
        refresh_interval       = "PT5M"
        data_unit              = "Percent"
        signal_definition_name = azurerm_monitor_health_model_signal_definition.metric.name

        evaluation_rule {
          unhealthy {
            operator  = "LessThan"
            threshold = 99
          }
        }
      }
    }

    azure_log_analytics {
      authentication_setting_name = azurerm_monitor_health_model_authentication_setting.test.name
      log_analytics_workspace_id  = azurerm_log_analytics_workspace.test.id

      signal {
        name                   = "heartbeat"
        display_name           = "Heartbeat"
        query_text             = "Heartbeat | summarize Total = count() by bin(TimeGenerated, 5m)"
        time_grain             = "PT5M"
        value_column_name      = "Total"
        refresh_interval       = "PT5M"
        data_unit              = "Count"
        signal_definition_name = azurerm_monitor_health_model_signal_definition.log.name

        evaluation_rule {
          unhealthy {
            operator  = "LessThan"
            threshold = 1
          }
        }
      }
    }

    azure_monitor_workspace {
      authentication_setting_name = azurerm_monitor_health_model_authentication_setting.test.name
      monitor_workspace_id        = azurerm_monitor_workspace.test.id

      signal {
        name                   = "prometheus-up"
        display_name           = "Up"
        query_text             = "up"
        time_grain             = "PT5M"
        refresh_interval       = "PT5M"
        data_unit              = "Count"
        signal_definition_name = azurerm_monitor_health_model_signal_definition.prometheus.name

        evaluation_rule {
          unhealthy {
            operator  = "LessThan"
            threshold = 1
          }
        }
      }
    }
  }
}
`, r.template(data), data.RandomInteger, "PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIxNiIgaGVpZ2h0PSIxNiI+PGNpcmNsZSBjeD0iOCIgY3k9IjgiIHI9IjciLz48L3N2Zz4=")
}

func (r MonitorHealthModelEntityResource) dynamicOperatorWithThreshold(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_monitor_health_model_entity" "test" {
  name            = "acctest-entity-%[2]d"
  health_model_id = azurerm_monitor_health_model.test.id

  signal_group {
    azure_resource {
      authentication_setting_name = azurerm_monitor_health_model_authentication_setting.test.name
      target_resource_id          = azurerm_storage_account.test.id

      signal {
        name             = "availability"
        metric_namespace = "Microsoft.Storage/storageAccounts"
        metric_name      = "Availability"
        aggregation_type = "Average"
        time_grain       = "PT5M"

        evaluation_rule {
          unhealthy {
            operator    = "Dynamic"
            sensitivity = "Medium"
            threshold   = 99
          }
        }
      }
    }
  }
}
`, r.template(data), data.RandomInteger)
}
