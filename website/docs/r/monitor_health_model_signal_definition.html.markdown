---
subcategory: "Monitor Health Models"
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_monitor_health_model_signal_definition"
description: |-
  Manages a Signal Definition within an Azure Monitor Health Model.
---

# azurerm_monitor_health_model_signal_definition

Manages a Signal Definition within an Azure Monitor Health Model.

A Signal Definition is a reusable signal template. Entities reference it by name through `signal_definition_name` instead of repeating the same metric, query and threshold configuration inline.

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
}

resource "azurerm_monitor_health_model_signal_definition" "example" {
  name             = "storage-availability"
  health_model_id  = azurerm_monitor_health_model.example.id
  display_name     = "Storage Account Availability"
  refresh_interval = "PT5M"
  data_unit        = "Percent"

  azure_resource_metric {
    metric_namespace = "Microsoft.Storage/storageAccounts"
    metric_name      = "Availability"
    aggregation_type = "Average"
    time_grain       = "PT5M"
  }

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
```

## Arguments Reference

The following arguments are supported:

* `name` - (Required) The name which should be used for this Signal Definition. Must be between 3 and 260 characters long, start and end with a letter or a number, and may only contain letters, numbers and hyphens. Changing this forces a new resource to be created.

* `health_model_id` - (Required) The ID of the Health Model in which this Signal Definition should exist. Changing this forces a new resource to be created.

* `evaluation_rule` - (Required) An `evaluation_rule` block as defined below.

* `azure_resource_metric` - (Optional) An `azure_resource_metric` block as defined below. Exactly one of `azure_resource_metric`, `log_analytics_query` or `prometheus_metrics_query` must be specified. Changing this forces a new resource to be created.

* `log_analytics_query` - (Optional) A `log_analytics_query` block as defined below. Exactly one of `azure_resource_metric`, `log_analytics_query` or `prometheus_metrics_query` must be specified. Changing this forces a new resource to be created.

* `prometheus_metrics_query` - (Optional) A `prometheus_metrics_query` block as defined below. Exactly one of `azure_resource_metric`, `log_analytics_query` or `prometheus_metrics_query` must be specified. Changing this forces a new resource to be created.

* `display_name` - (Optional) The display name of this Signal Definition.

* `refresh_interval` - (Optional) How often the signal is evaluated. Possible values are `PT1M`, `PT5M`, `PT10M`, `PT15M`, `PT30M`, `PT1H` and `PT2H`. Defaults to `PT1M`.

* `data_unit` - (Optional) The unit of the signal result, for example `Percent`, `Count`, `Bytes` or `MilliSeconds`.

* `tags` - (Optional) A mapping of tags which should be assigned to this Signal Definition.

---

An `azure_resource_metric` block supports the following:

* `metric_namespace` - (Required) The Azure Monitor metric namespace, for example `Microsoft.Storage/storageAccounts`.

* `metric_name` - (Required) The Azure Monitor metric name.

* `aggregation_type` - (Required) The aggregation to apply to the metric. Possible values are `None`, `Average`, `Count`, `Minimum`, `Maximum` and `Total`.

* `time_grain` - (Required) The time range the signal is evaluated over, as an ISO 8601 duration, for example `PT5M`.

~> **Note:** `time_grain` must be a grain the target metric actually supports, and `refresh_interval` must be less than or equal to `time_grain`. Neither constraint is caught at plan time - both are rejected by the resource provider on apply.

* `dimension_filter` - (Optional) A filter applied to the metric's dimensions, for example `ResponseType eq 'ServerBusyError'`.

---

A `log_analytics_query` block supports the following:

* `query_text` - (Required) The KQL query to evaluate.

* `time_grain` - (Optional) The time range the signal is evaluated over, as an ISO 8601 duration. If this is not specified the KQL query must define its own time range.

* `value_column_name` - (Optional) The numeric column in the result set that is compared against the thresholds. Defaults to the first column.

---

A `prometheus_metrics_query` block supports the following:

* `query_text` - (Required) The PromQL query to evaluate.

* `time_grain` - (Optional) The time range the signal is evaluated over, as an ISO 8601 duration.

---

An `evaluation_rule` block supports the following:

* `unhealthy` - (Required) An `unhealthy` block as defined below.

* `degraded` - (Optional) A `degraded` block as defined below.

---

An `unhealthy` or `degraded` block supports the following:

* `operator` - (Required) How the signal value is compared with the threshold. Possible values are `Equal`, `GreaterThan`, `GreaterThanOrEqual`, `LessThan`, `LessThanOrEqual`, `NotEqual` and `Dynamic`.

* `threshold` - (Optional) The threshold the signal value is compared against. This is required unless `operator` is `Dynamic`, where it is ignored.

* `sensitivity` - (Optional) The sensitivity of dynamic threshold detection. Possible values are `Low`, `Medium` and `High`. This can only be set when `operator` is `Dynamic`.

* `look_back_window` - (Optional) The historical window used to compute the dynamic threshold. Possible values are `PT5M`, `PT15M`, `PT30M` and `PT1H`. This can only be set when `operator` is `Dynamic`.

~> **Note:** `Dynamic` is only valid on the `unhealthy` rule.

## Attributes Reference

In addition to the Arguments listed above - the following Attributes are exported:

* `id` - The ID of the Signal Definition.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://developer.hashicorp.com/terraform/language/resources/configure#define-operation-timeouts) for certain actions:

* `create` - (Defaults to 30 minutes) Used when creating the Signal Definition.
* `read` - (Defaults to 5 minutes) Used when retrieving the Signal Definition.
* `update` - (Defaults to 30 minutes) Used when updating the Signal Definition.
* `delete` - (Defaults to 30 minutes) Used when deleting the Signal Definition.

## Import

Health Model Signal Definitions can be imported using the `resource id`, e.g.

```shell
terraform import azurerm_monitor_health_model_signal_definition.example /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/resourceGroup1/providers/Microsoft.CloudHealth/healthmodels/model1/signaldefinitions/definition1
```
