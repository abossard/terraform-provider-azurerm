provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "${var.prefix}-resources"
  location = var.location
}

resource "azurerm_storage_account" "example" {
  name                     = "${var.prefix}sa"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_tier             = "Standard"
  account_replication_type = "LRS"

  # The Health Model reads this account through its managed identity, so key-based and public access
  # are both pinned off. Leaving them at their defaults produces a perpetual diff in subscriptions
  # whose policies enforce these settings.
  public_network_access_enabled = false
  shared_access_key_enabled     = false
}

resource "azurerm_monitor_health_model" "example" {
  name                = "${var.prefix}-health-model"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location

  identity {
    type = "SystemAssigned"
  }
}

resource "azurerm_role_assignment" "example" {
  scope                = azurerm_resource_group.example.id
  role_definition_name = "Reader"
  principal_id         = azurerm_monitor_health_model.example.identity[0].principal_id
}

resource "azurerm_monitor_health_model_authentication_setting" "example" {
  name             = "${var.prefix}-authentication-setting"
  health_model_id  = azurerm_monitor_health_model.example.id
  display_name     = "System Assigned Identity"
  managed_identity = "SystemAssigned"
}

resource "azurerm_monitor_health_model_entity" "storage" {
  name                        = "${var.prefix}-storage-account"
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

  depends_on = [azurerm_role_assignment.example]
}

# The resource provider creates a root Entity named after the Health Model itself, so the
# Relationship below attaches the Storage Account Entity to it without declaring it here.
resource "azurerm_monitor_health_model_relationship" "storage" {
  name               = "${var.prefix}-workload-to-storage"
  health_model_id    = azurerm_monitor_health_model.example.id
  parent_entity_name = azurerm_monitor_health_model.example.name
  child_entity_name  = azurerm_monitor_health_model_entity.storage.name
  display_name       = "Workload depends on Storage Account"
}

resource "azurerm_monitor_health_model_signal_definition" "storage_availability" {
  name             = "${var.prefix}-storage-availability"
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

resource "azurerm_monitor_health_model_discovery_rule" "example" {
  name                        = "${var.prefix}-discover-resources"
  health_model_id             = azurerm_monitor_health_model.example.id
  display_name                = "Discover workload resources"
  authentication_setting_name = azurerm_monitor_health_model_authentication_setting.example.name

  resource_graph_query = "resources | where resourceGroup == '${azurerm_resource_group.example.name}' | project id"

  relationship_discovery_enabled = true
  recommended_signals_enabled    = true
  resource_health_signal_enabled = true

  depends_on = [azurerm_role_assignment.example]
}
