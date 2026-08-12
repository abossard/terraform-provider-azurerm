---
subcategory: "Monitor Health Models"
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_monitor_health_model_discovery_rule"
description: |-
  Manages a Discovery Rule within an Azure Monitor Health Model.
---

# azurerm_monitor_health_model_discovery_rule

Manages a Discovery Rule within an Azure Monitor Health Model.

A Discovery Rule populates a Health Model with Entities, and optionally Relationships, from an Azure Resource Graph query or an Application Insights topology.

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
  managed_identity = "SystemAssigned"
}

resource "azurerm_monitor_health_model_discovery_rule" "example" {
  name                        = "discover-workload-resources"
  health_model_id             = azurerm_monitor_health_model.example.id
  display_name                = "Discover workload resources"
  authentication_setting_name = azurerm_monitor_health_model_authentication_setting.example.name

  resource_graph_query = "resources | where resourceGroup == 'example-resources' | project id"

  relationship_discovery_enabled = true
  recommended_signals_enabled    = true
}
```

## Arguments Reference

The following arguments are supported:

* `name` - (Required) The name which should be used for this Discovery Rule. Must be between 3 and 260 characters long, start and end with a letter or a number, and may only contain letters, numbers and hyphens. Changing this forces a new resource to be created.

* `health_model_id` - (Required) The ID of the Health Model in which this Discovery Rule should exist. Changing this forces a new resource to be created.

* `authentication_setting_name` - (Required) The name of the `azurerm_monitor_health_model_authentication_setting` used to run the discovery query. The same Authentication Setting is assigned to every discovered Entity.

* `resource_graph_query` - (Optional) The Azure Resource Graph query, in KQL syntax, used to discover resources. The query must project a column named `id` containing the resource ID. Exactly one of `resource_graph_query` or `application_insights_id` must be specified. Changing this forces a new resource to be created.

* `application_insights_id` - (Optional) The ID of the Application Insights component whose topology is used to discover resources. Exactly one of `resource_graph_query` or `application_insights_id` must be specified. Changing this forces a new resource to be created.

* `display_name` - (Optional) The display name of this Discovery Rule.

* `relationship_discovery_enabled` - (Optional) Whether Relationships between discovered Entities should be created automatically using a set of built-in rules. Defaults to `true`.

~> **Note:** Relationships created by a Discovery Rule cannot be deleted manually.

* `recommended_signals_enabled` - (Optional) Whether all recommended signals should be added to discovered Entities. Defaults to `true`.

* `resource_health_signal_enabled` - (Optional) Whether a signal based on each discovered resource's Azure Resource Health availability state should be added automatically. Defaults to `true`.

## Attributes Reference

In addition to the Arguments listed above - the following Attributes are exported:

* `id` - The ID of the Discovery Rule.

* `entity_name` - The name of the Entity which represents this Discovery Rule. This Entity may take a few minutes to appear after the Discovery Rule is created.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://developer.hashicorp.com/terraform/language/resources/configure#define-operation-timeouts) for certain actions:

* `create` - (Defaults to 30 minutes) Used when creating the Discovery Rule.
* `read` - (Defaults to 5 minutes) Used when retrieving the Discovery Rule.
* `update` - (Defaults to 30 minutes) Used when updating the Discovery Rule.
* `delete` - (Defaults to 30 minutes) Used when deleting the Discovery Rule.

## Import

Health Model Discovery Rules can be imported using the `resource id`, e.g.

```shell
terraform import azurerm_monitor_health_model_discovery_rule.example /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/resourceGroup1/providers/Microsoft.CloudHealth/healthmodels/model1/discoveryrules/rule1
```
