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

type MonitorHealthModelDiscoveryRuleResource struct{}

func TestAccMonitorHealthModelDiscoveryRule_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model_discovery_rule", "test")
	r := MonitorHealthModelDiscoveryRuleResource{}
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

func TestAccMonitorHealthModelDiscoveryRule_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model_discovery_rule", "test")
	r := MonitorHealthModelDiscoveryRuleResource{}
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

func TestAccMonitorHealthModelDiscoveryRule_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model_discovery_rule", "test")
	r := MonitorHealthModelDiscoveryRuleResource{}
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

func TestAccMonitorHealthModelDiscoveryRule_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model_discovery_rule", "test")
	r := MonitorHealthModelDiscoveryRuleResource{}
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

// TestAccMonitorHealthModelDiscoveryRule_applicationInsightsTopology covers the second
// `DiscoveryRuleKind` discriminator value from the 2026-05-01-preview swagger.
func TestAccMonitorHealthModelDiscoveryRule_applicationInsightsTopology(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model_discovery_rule", "test")
	r := MonitorHealthModelDiscoveryRuleResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.applicationInsightsTopology(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (r MonitorHealthModelDiscoveryRuleResource) Exists(_ context.Context, _ *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.DiscoveryRuleID(state.ID)
	if err != nil {
		return nil, err
	}

	// TODO: replace with `clients.CloudHealth.DiscoveryRulesClient.Get(ctx, *id)` once a stable
	// `Microsoft.CloudHealth` API version ships and `hashicorp/go-azure-sdk` is vendored.
	return nil, fmt.Errorf("checking for %s: %w", id, errSDKUnavailable)
}

func (r MonitorHealthModelDiscoveryRuleResource) template(data acceptance.TestData) string {
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

resource "azurerm_monitor_health_model_authentication_setting" "test" {
  name             = "acctest-as-%[1]d"
  health_model_id  = azurerm_monitor_health_model.test.id
  managed_identity = "SystemAssigned"
}
`, data.RandomInteger, data.Locations.Primary)
}

func (r MonitorHealthModelDiscoveryRuleResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_monitor_health_model_discovery_rule" "test" {
  name                        = "acctest-dr-%d"
  health_model_id             = azurerm_monitor_health_model.test.id
  authentication_setting_name = azurerm_monitor_health_model_authentication_setting.test.name

  resource_graph_query = "resources | where resourceGroup == '${azurerm_resource_group.test.name}' | project id"
}
`, r.template(data), data.RandomInteger)
}

func (r MonitorHealthModelDiscoveryRuleResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_monitor_health_model_discovery_rule" "import" {
  name                        = azurerm_monitor_health_model_discovery_rule.test.name
  health_model_id             = azurerm_monitor_health_model_discovery_rule.test.health_model_id
  authentication_setting_name = azurerm_monitor_health_model_discovery_rule.test.authentication_setting_name
  resource_graph_query        = azurerm_monitor_health_model_discovery_rule.test.resource_graph_query
}
`, r.basic(data))
}

func (r MonitorHealthModelDiscoveryRuleResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_monitor_health_model_discovery_rule" "test" {
  name                        = "acctest-dr-%d"
  health_model_id             = azurerm_monitor_health_model.test.id
  display_name                = "Discover workload resources"
  authentication_setting_name = azurerm_monitor_health_model_authentication_setting.test.name

  resource_graph_query = "resources | where resourceGroup == '${azurerm_resource_group.test.name}' | project id"

  relationship_discovery_enabled = true
  recommended_signals_enabled    = true
  resource_health_signal_enabled = false
}
`, r.template(data), data.RandomInteger)
}

func (r MonitorHealthModelDiscoveryRuleResource) applicationInsightsTopology(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_log_analytics_workspace" "test" {
  name                = "acctest-law-%[2]d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  sku                 = "PerGB2018"
  retention_in_days   = 30
}

resource "azurerm_application_insights" "test" {
  name                = "acctest-appi-%[2]d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  workspace_id        = azurerm_log_analytics_workspace.test.id
  application_type    = "web"
}

resource "azurerm_monitor_health_model_discovery_rule" "test" {
  name                        = "acctest-dr-%[2]d"
  health_model_id             = azurerm_monitor_health_model.test.id
  authentication_setting_name = azurerm_monitor_health_model_authentication_setting.test.name
  application_insights_id     = azurerm_application_insights.test.id

  relationship_discovery_enabled = true
  recommended_signals_enabled    = true
}
`, r.template(data), data.RandomInteger)
}
