// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cloudhealth

// Resource ID helpers remain generated in-provider until CloudHealth is available in the vendored
// `hashicorp/go-azure-sdk`, using the legacy generator documented in
// `contributing/topics/guide-resource-ids.md`.
//
// The ARM path segments match the `2026-11-01` stable swagger at
// `Azure/azure-rest-api-specs`: `specification/cloudhealth/resource-manager/Microsoft.CloudHealth/CloudHealth/stable/2026-11-01/cloudhealth.json`.

//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=HealthModel -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.CloudHealth/healthmodels/model1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=AuthenticationSetting -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.CloudHealth/healthmodels/model1/authenticationsettings/setting1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=DiscoveryRule -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.CloudHealth/healthmodels/model1/discoveryrules/rule1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=Entity -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.CloudHealth/healthmodels/model1/entities/entity1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=Relationship -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.CloudHealth/healthmodels/model1/relationships/relationship1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=SignalDefinition -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.CloudHealth/healthmodels/model1/signaldefinitions/definition1
