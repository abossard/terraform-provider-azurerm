---
subcategory: "Monitor Health Models"
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_monitor_health_model_relationship"
description: |-
  Manages a Relationship between two Entities within an Azure Monitor Health Model.
---

# azurerm_monitor_health_model_relationship

Manages a Relationship between two Entities within an Azure Monitor Health Model.

A Relationship is an edge in the health graph. The child Entity's health state rolls up into the parent Entity according to the parent's `dependency` configuration.

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

# The resource provider creates a root Entity whose name matches the Health Model, so this
# Relationship can reference it without declaring it. Declaring an Entity with that name would fail,
# because it already exists - import it instead if you need to manage its properties.
resource "azurerm_monitor_health_model_entity" "data_tier" {
  name            = "example-data-tier"
  health_model_id = azurerm_monitor_health_model.example.id
  display_name    = "Data Tier"

  signal_group {
    dependency {
      aggregation_type = "WorstOf"
    }
  }
}

resource "azurerm_monitor_health_model_relationship" "example" {
  name               = "r-workload-data-tier"
  health_model_id    = azurerm_monitor_health_model.example.id
  parent_entity_name = azurerm_monitor_health_model.example.name
  child_entity_name  = azurerm_monitor_health_model_entity.data_tier.name
  display_name       = "Workload depends on Data Tier"
}
```

## Arguments Reference

The following arguments are supported:

* `name` - (Required) The name which should be used for this Relationship. Must be between 3 and 260 characters long, start and end with a letter or a number, and may only contain letters, numbers and hyphens. Changing this forces a new resource to be created.

~> **Note:** Prefer a deterministic name such as `r-<parent>-<child>` over a generated GUID, so the Relationship stays stable across applies.

* `health_model_id` - (Required) The ID of the Health Model in which this Relationship should exist. Changing this forces a new resource to be created.

* `parent_entity_name` - (Required) The name of the parent Entity. Changing this forces a new resource to be created.

* `child_entity_name` - (Required) The name of the child Entity. Changing this forces a new resource to be created.

~> **Note:** A parent and child Entity pair may only be connected by one Relationship.

* `display_name` - (Optional) The display name of this Relationship.

* `tags` - (Optional) A mapping of tags which should be assigned to this Relationship.

## Attributes Reference

In addition to the Arguments listed above - the following Attributes are exported:

* `id` - The ID of the Relationship.

* `discovered_by` - The name of the `azurerm_monitor_health_model_discovery_rule` that discovered this Relationship, if any. Relationships that were discovered cannot be deleted manually.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://developer.hashicorp.com/terraform/language/resources/configure#define-operation-timeouts) for certain actions:

* `create` - (Defaults to 30 minutes) Used when creating the Relationship.
* `read` - (Defaults to 5 minutes) Used when retrieving the Relationship.
* `update` - (Defaults to 30 minutes) Used when updating the Relationship.
* `delete` - (Defaults to 30 minutes) Used when deleting the Relationship.

## Import

Health Model Relationships can be imported using the `resource id`, e.g.

```shell
terraform import azurerm_monitor_health_model_relationship.example /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/resourceGroup1/providers/Microsoft.CloudHealth/healthmodels/model1/relationships/relationship1
```
