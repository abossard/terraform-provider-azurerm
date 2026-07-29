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

type MonitorHealthModelAuthenticationSettingResource struct{}

func TestAccMonitorHealthModelAuthenticationSetting_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model_authentication_setting", "test")
	r := MonitorHealthModelAuthenticationSettingResource{}
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

func TestAccMonitorHealthModelAuthenticationSetting_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model_authentication_setting", "test")
	r := MonitorHealthModelAuthenticationSettingResource{}
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

func TestAccMonitorHealthModelAuthenticationSetting_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model_authentication_setting", "test")
	r := MonitorHealthModelAuthenticationSettingResource{}
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

func TestAccMonitorHealthModelAuthenticationSetting_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_health_model_authentication_setting", "test")
	r := MonitorHealthModelAuthenticationSettingResource{}
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

func (r MonitorHealthModelAuthenticationSettingResource) Exists(_ context.Context, _ *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.AuthenticationSettingID(state.ID)
	if err != nil {
		return nil, err
	}

	// TODO: replace with `clients.CloudHealth.AuthenticationSettingsClient.Get(ctx, *id)` once a stable
	// `Microsoft.CloudHealth` API version ships and `hashicorp/go-azure-sdk` is vendored.
	return nil, fmt.Errorf("checking for %s: %w", id, errSDKUnavailable)
}

func (r MonitorHealthModelAuthenticationSettingResource) template(data acceptance.TestData) string {
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

func (r MonitorHealthModelAuthenticationSettingResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_monitor_health_model_authentication_setting" "test" {
  name             = "acctest-as-%d"
  health_model_id  = azurerm_monitor_health_model.test.id
  managed_identity = "SystemAssigned"
}
`, r.template(data), data.RandomInteger)
}

func (r MonitorHealthModelAuthenticationSettingResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_monitor_health_model_authentication_setting" "import" {
  name             = azurerm_monitor_health_model_authentication_setting.test.name
  health_model_id  = azurerm_monitor_health_model_authentication_setting.test.health_model_id
  managed_identity = azurerm_monitor_health_model_authentication_setting.test.managed_identity
}
`, r.basic(data))
}

func (r MonitorHealthModelAuthenticationSettingResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_user_assigned_identity" "test" {
  name                = "acctest-uai-%[2]d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
}

resource "azurerm_monitor_health_model_authentication_setting" "test" {
  name             = "acctest-as-%[2]d"
  health_model_id  = azurerm_monitor_health_model.test.id
  display_name     = "Acceptance Test Authentication Setting"
  managed_identity = azurerm_user_assigned_identity.test.id
}
`, r.template(data), data.RandomInteger)
}
