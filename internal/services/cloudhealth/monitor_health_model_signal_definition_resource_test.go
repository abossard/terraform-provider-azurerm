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
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/cloudhealth/sdk/2026-11-01/signaldefinitions"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
)

type MonitorHealthModelSignalDefinitionResource struct{}

func TestAccMonitorHealthModelSignalDefinition_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model_signal_definition", "test")
	r := MonitorHealthModelSignalDefinitionResource{}
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

func TestAccMonitorHealthModelSignalDefinition_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model_signal_definition", "test")
	r := MonitorHealthModelSignalDefinitionResource{}
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

func TestAccMonitorHealthModelSignalDefinition_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model_signal_definition", "test")
	r := MonitorHealthModelSignalDefinitionResource{}
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

func TestAccMonitorHealthModelSignalDefinition_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model_signal_definition", "test")
	r := MonitorHealthModelSignalDefinitionResource{}
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

// TestAccMonitorHealthModelSignalDefinition_logAnalyticsQuery covers the second of the three writable
// `signalKind` discriminator values from the 2026-11-01 swagger.
func TestAccMonitorHealthModelSignalDefinition_logAnalyticsQuery(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model_signal_definition", "test")
	r := MonitorHealthModelSignalDefinitionResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.logAnalyticsQuery(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccMonitorHealthModelSignalDefinition_prometheusMetricsQuery(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model_signal_definition", "test")
	r := MonitorHealthModelSignalDefinitionResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.prometheusMetricsQuery(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccMonitorHealthModelSignalDefinition_dynamicOperator(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model_signal_definition", "test")
	r := MonitorHealthModelSignalDefinitionResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.dynamicOperator(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccMonitorHealthModelSignalDefinition_dynamicOperatorWithThreshold(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model_signal_definition", "test")
	r := MonitorHealthModelSignalDefinitionResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config:      r.dynamicOperatorWithThreshold(data),
			PlanOnly:    true,
			ExpectError: regexp.MustCompile("`threshold` cannot be set in `evaluation_rule.unhealthy` when `operator` is `Dynamic`"),
		},
	})
}

func TestAccMonitorHealthModelSignalDefinition_staticOperatorWithSensitivity(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model_signal_definition", "test")
	r := MonitorHealthModelSignalDefinitionResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config:      r.staticOperatorWithSensitivity(data),
			PlanOnly:    true,
			ExpectError: regexp.MustCompile("`sensitivity` cannot be set in `evaluation_rule.unhealthy` when `operator` is `LessThan`"),
		},
	})
}

func (r MonitorHealthModelSignalDefinitionResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := signaldefinitions.ParseSignaldefinitionID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.CloudHealth.SignalDefinitions.Get(ctx, *id)
	if err != nil {
		if response.WasNotFound(resp.HttpResponse) {
			return pointer.To(false), nil
		}
		return nil, fmt.Errorf("retrieving %s: %+v", *id, err)
	}

	return pointer.To(resp.Model != nil), nil
}

func (r MonitorHealthModelSignalDefinitionResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctest-rg-%[1]d"
  location = "%[2]s"
}

resource "azurerm_monitor_health_model" "test" {
  name                = "acctest-hm-%[1]d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location

  identity {
    type = "SystemAssigned"
  }
}
`, data.RandomInteger, data.Locations.Primary)
}

func (r MonitorHealthModelSignalDefinitionResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_monitor_health_model_signal_definition" "test" {
  name            = "acctest-sd-%d"
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
`, r.template(data), data.RandomInteger)
}

func (r MonitorHealthModelSignalDefinitionResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_monitor_health_model_signal_definition" "import" {
  name            = azurerm_monitor_health_model_signal_definition.test.name
  health_model_id = azurerm_monitor_health_model_signal_definition.test.health_model_id

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
`, r.basic(data))
}

func (r MonitorHealthModelSignalDefinitionResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_monitor_health_model_signal_definition" "test" {
  name             = "acctest-sd-%d"
  health_model_id  = azurerm_monitor_health_model.test.id
  display_name     = "Storage Account Availability"
  refresh_interval = "PT5M"
  data_unit        = "Percent"

  azure_resource_metric {
    metric_namespace = "Microsoft.Storage/storageAccounts"
    metric_name      = "Transactions"
    aggregation_type = "Total"
    time_grain       = "PT5M"
    dimension_filter = "ResponseType eq 'ServerBusyError'"
  }

  evaluation_rule {
    degraded {
      operator  = "GreaterThan"
      threshold = 10
    }

    unhealthy {
      operator  = "GreaterThan"
      threshold = 100
    }
  }

  tags = {
    tier = "data"
  }
}
`, r.template(data), data.RandomInteger)
}

func (r MonitorHealthModelSignalDefinitionResource) logAnalyticsQuery(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_monitor_health_model_signal_definition" "test" {
  name             = "acctest-sd-%d"
  health_model_id  = azurerm_monitor_health_model.test.id
  display_name     = "Failed Requests"
  refresh_interval = "PT5M"
  data_unit        = "Count"

  log_analytics_query {
    query_text        = "AppRequests | where Success == false | summarize Failures = count()"
    time_grain        = "PT5M"
    value_column_name = "Failures"
  }

  evaluation_rule {
    unhealthy {
      operator  = "GreaterThan"
      threshold = 25
    }
  }
}
`, r.template(data), data.RandomInteger)
}

func (r MonitorHealthModelSignalDefinitionResource) dynamicOperatorWithThreshold(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_monitor_health_model_signal_definition" "test" {
  name            = "acctest-sd-%d"
  health_model_id = azurerm_monitor_health_model.test.id

  azure_resource_metric {
    metric_namespace = "Microsoft.Storage/storageAccounts"
    metric_name      = "Availability"
    aggregation_type = "Average"
    time_grain       = "PT5M"
  }

  evaluation_rule {
    unhealthy {
      operator    = "Dynamic"
      sensitivity = "Medium"
      threshold   = 99
    }
  }
}
`, r.template(data), data.RandomInteger)
}

func (r MonitorHealthModelSignalDefinitionResource) staticOperatorWithSensitivity(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_monitor_health_model_signal_definition" "test" {
  name            = "acctest-sd-%d"
  health_model_id = azurerm_monitor_health_model.test.id

  azure_resource_metric {
    metric_namespace = "Microsoft.Storage/storageAccounts"
    metric_name      = "Availability"
    aggregation_type = "Average"
    time_grain       = "PT5M"
  }

  evaluation_rule {
    unhealthy {
      operator    = "LessThan"
      threshold   = 99
      sensitivity = "Medium"
    }
  }
}
`, r.template(data), data.RandomInteger)
}

func (r MonitorHealthModelSignalDefinitionResource) dynamicOperator(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_monitor_health_model_signal_definition" "test" {
  name            = "acctest-sd-%d"
  health_model_id = azurerm_monitor_health_model.test.id

  azure_resource_metric {
    metric_namespace = "Microsoft.Storage/storageAccounts"
    metric_name      = "Availability"
    aggregation_type = "Average"
    time_grain       = "PT5M"
  }

  evaluation_rule {
    unhealthy {
      operator    = "Dynamic"
      sensitivity = "Medium"
    }
  }
}
`, r.template(data), data.RandomInteger)
}

func (r MonitorHealthModelSignalDefinitionResource) prometheusMetricsQuery(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_monitor_health_model_signal_definition" "test" {
  name             = "acctest-sd-%d"
  health_model_id  = azurerm_monitor_health_model.test.id
  display_name     = "Prometheus Up"
  refresh_interval = "PT5M"
  data_unit        = "Count"

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
`, r.template(data), data.RandomInteger)
}
