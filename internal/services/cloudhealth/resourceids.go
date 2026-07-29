// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cloudhealth

// NOTE: `Microsoft.CloudHealth` only publishes preview API versions, so `hashicorp/pandora` has not
// generated a `hashicorp/go-azure-sdk` package for it and the Resource ID parsers/validators that
// would normally come from that package do not exist. Until a stable API version ships these IDs are
// generated in-provider using the legacy generator documented in
// `contributing/topics/guide-resource-ids.md`, which that guide reserves for exceptional cases.
//
// The ARM path segments below are taken verbatim from the `2026-05-01-preview` swagger at
// `Azure/azure-rest-api-specs`: `specification/cloudhealth/resource-manager/Microsoft.CloudHealth/CloudHealth/preview/2026-05-01-preview/cloudhealth.json`.
// Path segments do not change across a preview -> stable transition, so these IDs remain valid at GA.

//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=HealthModel -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.CloudHealth/healthmodels/model1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=AuthenticationSetting -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.CloudHealth/healthmodels/model1/authenticationsettings/setting1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=DiscoveryRule -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.CloudHealth/healthmodels/model1/discoveryrules/rule1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=Entity -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.CloudHealth/healthmodels/model1/entities/entity1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=Relationship -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.CloudHealth/healthmodels/model1/relationships/relationship1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=SignalDefinition -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.CloudHealth/healthmodels/model1/signaldefinitions/definition1
