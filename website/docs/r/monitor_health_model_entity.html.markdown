---
subcategory: "Monitor Health Models"
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_monitor_health_model_entity"
description: |-
  Manages an Entity within an Azure Monitor Health Model.
---

# azurerm_monitor_health_model_entity

Manages an Entity within an Azure Monitor Health Model.

An Entity is a node in the health graph. A leaf Entity derives its health from signals evaluated against an Azure resource, a Log Analytics Workspace or an Azure Monitor Workspace. A grouping Entity derives its health from its child Entities.

~> **Note:** The CloudHealth resource provider creates a root Entity automatically, named after the Health Model. Because it already exists, declaring an Entity with that name fails with a "resource already exists" error. To manage it with Terraform, write the `resource` block and then `terraform import` it onto the existing root rather than letting Terraform create it, e.g. `terraform import azurerm_monitor_health_model_entity.root /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/example-resources/providers/Microsoft.CloudHealth/healthmodels/example-model/entities/example-model`. Do not give a second Entity a different name intending it as the root: that creates a competing root rather than replacing the generated one.

## Example Usage

```hcl
resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "North Europe"
}

resource "azurerm_storage_account" "example" {
  name                     = "examplestorageacc"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
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

resource "azurerm_monitor_health_model_entity" "example" {
  name                        = "example-storage-account"
  health_model_id             = azurerm_monitor_health_model.example.id
  display_name                = "Storage Account"
  health_objective_percentage = 99.9

  canvas_position {
    x = 120
    y = 480
  }

  signal_group {
    azure_resource {
      authentication_setting_name = azurerm_monitor_health_model_authentication_setting.example.name
      target_resource_id          = azurerm_storage_account.example.id
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
    }
  }
}
```

## Arguments Reference

The following arguments are supported:

* `name` - (Required) The name which should be used for this Entity. Must be between 3 and 260 characters long, start and end with a letter or a number, and may only contain letters, numbers and hyphens. Changing this forces a new resource to be created.

* `health_model_id` - (Required) The ID of the Health Model in which this Entity should exist. Changing this forces a new resource to be created.

* `display_name` - (Optional) The display name of this Entity.

* `health_objective_percentage` - (Optional) The percentage of time this Entity is expected to be healthy. Possible values are between `0` and `100`.

* `impact` - (Optional) How this Entity's health state propagates to its parents. Possible values are `Standard`, `Limited` and `Suppressed`. Defaults to `Standard`.

* `canvas_position` - (Optional) A `canvas_position` block as defined below.

~> **Note:** Without `canvas_position`, the portal's health model graph stacks every Entity in a single vertical column.

* `icon` - (Optional) An `icon` block as defined below.

* `signal_group` - (Optional) A `signal_group` block as defined below.

* `alert` - (Optional) An `alert` block as defined below.

* `tags` - (Optional) A mapping of tags which should be assigned to this Entity.

---

A `canvas_position` block supports the following:

* `x` - (Required) The X coordinate of this Entity on the health model canvas.

* `y` - (Required) The Y coordinate of this Entity on the health model canvas.

---

An `icon` block supports the following:

* `name` - (Required) The name of the built-in icon to use, or `Custom` to use `custom_data_base64`.

* `custom_data_base64` - (Optional) Base64 encoded SVG data to use as the icon. Setting this overrides the built-in icon.

---

A `signal_group` block supports the following:

* `azure_resource` - (Optional) An `azure_resource` block as defined below.

* `azure_log_analytics` - (Optional) An `azure_log_analytics` block as defined below.

* `azure_monitor_workspace` - (Optional) An `azure_monitor_workspace` block as defined below.

* `dependency` - (Optional) A `dependency` block as defined below.

---

An `azure_resource` block supports the following:

* `authentication_setting_name` - (Required) The name of the `azurerm_monitor_health_model_authentication_setting` used to query this data source.

* `target_resource_id` - (Required) The ID of the Azure resource this Entity represents.

* `resource_health_enabled` - (Optional) Whether a signal based on the resource's Azure Resource Health availability state should be added automatically. Defaults to `true`.

* `signal` - (Optional) One or more `signal` blocks as defined below. A maximum of 50 signals may be assigned.

---

A `signal` block within `azure_resource` supports the following:

* `name` - (Required) The name of this signal, unique within the Entity.

* `metric_namespace` - (Required) The Azure Monitor metric namespace, for example `Microsoft.Storage/storageAccounts`.

* `metric_name` - (Required) The Azure Monitor metric name.

* `aggregation_type` - (Required) The aggregation to apply to the metric. Possible values are `None`, `Average`, `Count`, `Minimum`, `Maximum` and `Total`.

* `time_grain` - (Required) The time range the signal is evaluated over, as an ISO 8601 duration, for example `PT5M`.

~> **Note:** `time_grain` must be a grain the target metric supports, and `refresh_interval` must be less than or equal to `time_grain`. Terraform cannot check either constraint at plan time; the resource provider rejects both on apply.

* `evaluation_rule` - (Required) An `evaluation_rule` block as defined below.

* `display_name` - (Optional) The display name of this signal.

* `dimension_filter` - (Optional) A filter applied to the metric's dimensions, for example `ResponseType eq 'ServerBusyError'`.

* `refresh_interval` - (Optional) How often the signal is evaluated. Possible values are `PT1M`, `PT5M`, `PT10M`, `PT15M`, `PT30M`, `PT1H` and `PT2H`. Defaults to `PT1M`.

* `data_unit` - (Optional) The unit of the signal result, for example `Percent`, `Count`, `Bytes` or `MilliSeconds`.

* `signal_definition_name` - (Optional) The name of an `azurerm_monitor_health_model_signal_definition` that supplies default values for this signal.

---

An `azure_log_analytics` block supports the following:

* `authentication_setting_name` - (Required) The name of the `azurerm_monitor_health_model_authentication_setting` used to query this data source.

* `log_analytics_workspace_id` - (Required) The ID of the Log Analytics Workspace to query.

* `signal` - (Optional) One or more `signal` blocks as defined below. A maximum of 50 signals may be assigned.

---

A `signal` block within `azure_log_analytics` supports the following:

* `name` - (Required) The name of this signal, unique within the Entity.

* `query_text` - (Required) The KQL query to evaluate.

* `evaluation_rule` - (Required) An `evaluation_rule` block as defined below.

* `time_grain` - (Optional) The time range the signal is evaluated over, as an ISO 8601 duration. If this is not specified the KQL query must define its own time range.

* `value_column_name` - (Optional) The numeric column in the result set that is compared against the thresholds. Defaults to the first column.

* `display_name` - (Optional) The display name of this signal.

* `refresh_interval` - (Optional) How often the signal is evaluated. Possible values are `PT1M`, `PT5M`, `PT10M`, `PT15M`, `PT30M`, `PT1H` and `PT2H`. Defaults to `PT1M`.

* `data_unit` - (Optional) The unit of the signal result.

* `signal_definition_name` - (Optional) The name of an `azurerm_monitor_health_model_signal_definition` that supplies default values for this signal.

---

An `azure_monitor_workspace` block supports the following:

* `authentication_setting_name` - (Required) The name of the `azurerm_monitor_health_model_authentication_setting` used to query this data source.

* `monitor_workspace_id` - (Required) The ID of the Azure Monitor Workspace to query.

* `signal` - (Optional) One or more `signal` blocks as defined below. A maximum of 50 signals may be assigned.

---

A `signal` block within `azure_monitor_workspace` supports the following:

* `name` - (Required) The name of this signal, unique within the Entity.

* `query_text` - (Required) The PromQL query to evaluate.

* `evaluation_rule` - (Required) An `evaluation_rule` block as defined below.

* `time_grain` - (Optional) The time range the signal is evaluated over, as an ISO 8601 duration.

* `display_name` - (Optional) The display name of this signal.

* `refresh_interval` - (Optional) How often the signal is evaluated. Possible values are `PT1M`, `PT5M`, `PT10M`, `PT15M`, `PT30M`, `PT1H` and `PT2H`. Defaults to `PT1M`.

* `data_unit` - (Optional) The unit of the signal result.

* `signal_definition_name` - (Optional) The name of an `azurerm_monitor_health_model_signal_definition` that supplies default values for this signal.

---

An `evaluation_rule` block supports the following:

* `unhealthy` - (Required) An `unhealthy` block as defined below.

* `degraded` - (Optional) A `degraded` block as defined below.

---

An `unhealthy` or `degraded` block supports the following:

* `operator` - (Required) How the signal value is compared with the threshold. Possible values are `Equal`, `GreaterThan`, `GreaterThanOrEqual`, `LessThan`, `LessThanOrEqual`, `NotEqual` and `Dynamic`.

* `threshold` - (Optional) The threshold the signal value is compared against. This is required unless `operator` is `Dynamic`, where it is ignored.

* `sensitivity` - (Optional) The sensitivity of dynamic threshold detection. Possible values are `Low`, `Medium` and `High`. This can only be set when `operator` is `Dynamic`.

~> **Note:** `Dynamic` is only valid on the `unhealthy` rule.

---

A `dependency` block supports the following:

* `aggregation_type` - (Required) How the health states of child Entities roll up into this Entity. Possible values are `WorstOf`, `BestOf`, `MinHealthy` and `MaxNotHealthy`.

* `unit` - (Optional) The unit of the aggregation thresholds. Possible values are `Absolute` and `Percentage`. This is required when `aggregation_type` is `MinHealthy` or `MaxNotHealthy` and cannot be set for `WorstOf` or `BestOf`.

* `unhealthy_threshold` - (Optional) The threshold at which this Entity becomes unhealthy. This is required when `aggregation_type` is `MinHealthy` or `MaxNotHealthy` and cannot be set for `WorstOf` or `BestOf`.

* `degraded_threshold` - (Optional) The threshold at which this Entity becomes degraded. If this is not specified the Entity transitions directly from healthy to unhealthy. This cannot be set for `WorstOf` or `BestOf`.

* `ignore_unknown_enabled` - (Optional) Whether children with an `Unknown` health state are excluded from the aggregation. Defaults to `true`. This has no effect for `WorstOf` or `BestOf`.

---

An `alert` block supports the following:

* `unhealthy` - (Optional) An `unhealthy` block as defined below.

* `degraded` - (Optional) A `degraded` block as defined below.

---

An `unhealthy` or `degraded` block within `alert` supports the following:

* `severity` - (Required) The severity of the triggered alert. Possible values are `Sev0`, `Sev1`, `Sev2`, `Sev3` and `Sev4`.

* `description` - (Optional) The description of the alert rule.

* `action_group_ids` - (Optional) Between 1 and 5 Action Group IDs to notify when the alert triggers.

## Attributes Reference

In addition to the Arguments listed above - the following Attributes are exported:

* `id` - The ID of the Entity.

* `health_state` - The current health state of this Entity.

* `discovered_by` - The name of the `azurerm_monitor_health_model_discovery_rule` that discovered this Entity, if any. Entities that were discovered cannot be deleted manually.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://developer.hashicorp.com/terraform/language/resources/configure#define-operation-timeouts) for certain actions:

* `create` - (Defaults to 30 minutes) Used when creating the Entity.
* `read` - (Defaults to 5 minutes) Used when retrieving the Entity.
* `update` - (Defaults to 30 minutes) Used when updating the Entity.
* `delete` - (Defaults to 30 minutes) Used when deleting the Entity.

## Import

Health Model Entities can be imported using the `resource id`, e.g.

```shell
terraform import azurerm_monitor_health_model_entity.example /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/resourceGroup1/providers/Microsoft.CloudHealth/healthmodels/model1/entities/entity1
```
