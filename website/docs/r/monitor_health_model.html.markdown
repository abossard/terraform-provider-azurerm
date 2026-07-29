---
subcategory: "Monitor Health Models"
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_monitor_health_model"
description: |-
  Manages an Azure Monitor Health Model.
---

# azurerm_monitor_health_model

Manages an Azure Monitor Health Model.

## Example Usage

```hcl
resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_monitor_health_model" "example" {
  name                = "example-health-model"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location

  identity {
    type = "SystemAssigned"
  }

  tags = {
    environment = "production"
  }
}
```

## Arguments Reference

The following arguments are supported:

* `name` - (Required) The name which should be used for this Health Model. Must be between 3 and 44 characters long, start with a letter, end with a letter or a number, and may only contain letters, numbers and hyphens. Changing this forces a new resource to be created.

* `resource_group_name` - (Required) The name of the Resource Group in which the Health Model should exist. Changing this forces a new resource to be created.

* `location` - (Required) The Azure Region where the Health Model should exist. Changing this forces a new resource to be created.

~> **Note:** A Health Model can be created in a different region from the resources it monitors. Use `az provider show --namespace Microsoft.CloudHealth --query "resourceTypes[?resourceType=='healthmodels'].locations[]"` to list the supported regions rather than assuming the Resource Group's region is supported.

* `identity` - (Optional) An `identity` block as defined below.

~> **Note:** The identity assigned to the Health Model requires the `Reader` role over every scope it monitors. `Monitoring Reader` is not sufficient - signals evaluate to `Unknown` without `Reader`.

* `tags` - (Optional) A mapping of tags which should be assigned to the Health Model.

---

An `identity` block supports the following:

* `type` - (Required) The type of Managed Service Identity that should be assigned to this Health Model. Possible values are `SystemAssigned`, `UserAssigned` and `SystemAssigned, UserAssigned`.

* `identity_ids` - (Optional) A list of User Assigned Managed Identity IDs to be assigned to this Health Model. This is required when `type` is set to `UserAssigned` or `SystemAssigned, UserAssigned`.

## Attributes Reference

In addition to the Arguments listed above - the following Attributes are exported:

* `id` - The ID of the Health Model.

* `identity` - An `identity` block as defined below.

---

An `identity` block exports the following:

* `principal_id` - The Principal ID of the System Assigned Managed Service Identity that is configured on this Health Model.

* `tenant_id` - The Tenant ID of the System Assigned Managed Service Identity that is configured on this Health Model.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://developer.hashicorp.com/terraform/language/resources/configure#define-operation-timeouts) for certain actions:

* `create` - (Defaults to 30 minutes) Used when creating the Health Model.
* `read` - (Defaults to 5 minutes) Used when retrieving the Health Model.
* `update` - (Defaults to 30 minutes) Used when updating the Health Model.
* `delete` - (Defaults to 30 minutes) Used when deleting the Health Model.

## Import

Health Models can be imported using the `resource id`, e.g.

```shell
terraform import azurerm_monitor_health_model.example /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/resourceGroup1/providers/Microsoft.CloudHealth/healthmodels/model1
```
