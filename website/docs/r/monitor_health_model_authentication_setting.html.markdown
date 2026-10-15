---
subcategory: "Monitor Health Models"
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_monitor_health_model_authentication_setting"
description: |-
  Manages an Authentication Setting within an Azure Monitor Health Model.
---

# azurerm_monitor_health_model_authentication_setting

Manages an Authentication Setting within an Azure Monitor Health Model.

An Authentication Setting tells the Health Model which Managed Identity to use when it queries the data sources behind an Entity's signals.

## Example Usage

```hcl
resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "North Europe"
}

resource "azurerm_monitor_health_model" "example" {
  name                = "example-health-model"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location

  identity {
    type = "SystemAssigned"
  }
}

resource "azurerm_monitor_health_model_authentication_setting" "example" {
  name             = "example-authentication-setting"
  health_model_id  = azurerm_monitor_health_model.example.id
  display_name     = "System Assigned Identity"
  managed_identity = "SystemAssigned"
}
```

## Arguments Reference

The following arguments are supported:

* `name` - (Required) The name which should be used for this Authentication Setting. Must be between 3 and 260 characters long, start and end with a letter or a number, and may only contain letters, numbers and hyphens. Changing this forces a new resource to be created.

* `health_model_id` - (Required) The ID of the Health Model in which this Authentication Setting should exist. Changing this forces a new resource to be created.

* `managed_identity` - (Required) The Managed Identity that the Health Model uses when querying data sources. Use `SystemAssigned` for the Health Model's System Assigned Managed Identity. Otherwise, use the ID of a User Assigned Managed Identity assigned to the Health Model. Must be between 1 and 500 characters long.

* `display_name` - (Optional) The display name of this Authentication Setting. Must be between 1 and 260 characters long.

## Attributes Reference

In addition to the Arguments listed above - the following Attributes are exported:

* `id` - The ID of the Authentication Setting.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://developer.hashicorp.com/terraform/language/resources/configure#define-operation-timeouts) for certain actions:

* `create` - (Defaults to 30 minutes) Used when creating the Authentication Setting.
* `read` - (Defaults to 5 minutes) Used when retrieving the Authentication Setting.
* `update` - (Defaults to 30 minutes) Used when updating the Authentication Setting.
* `delete` - (Defaults to 30 minutes) Used when deleting the Authentication Setting.

## Import

Health Model Authentication Settings can be imported using the `resource id`, e.g.

```shell
terraform import azurerm_monitor_health_model_authentication_setting.example /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/resourceGroup1/providers/Microsoft.CloudHealth/healthmodels/model1/authenticationsettings/setting1
```
