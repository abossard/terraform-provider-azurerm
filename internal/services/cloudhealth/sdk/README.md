# `Microsoft.CloudHealth` SDK

This directory contains a generated Azure Resource Manager client for `Microsoft.CloudHealth`, API
version `2026-11-01` (stable).

## Why it lives here

`github.com/hashicorp/go-azure-sdk/resource-manager` does not yet include `Microsoft.CloudHealth`.
The provider resources use this local SDK until the upstream package is available.

## Generated layout

`pandora`'s `generator-go-sdk` generates the service packages. When `go-azure-sdk` includes
`Microsoft.CloudHealth`, replace the local packages with imports from:

```text
github.com/hashicorp/go-azure-sdk/resource-manager/cloudhealth/2026-11-01/<package>
```

- Package files follow the generated layout: `client.go`, `version.go`, `constants.go`, `id_*.go`,
  `method_*.go`, `model_*.go`, `predicates.go`.
- Files use the Microsoft/MIT header from generated `go-azure-sdk` code.
- SDK packages must not import `github.com/hashicorp/terraform-provider-azurerm`.
- Discriminated types (`AuthenticationSettingProperties`, `SignalInstanceProperties`,
  `SignalDefinitionProperties`, `DiscoveryRuleSpecification`) use the generated interface +
  `Base...Impl` + `Raw...Impl` + `Unmarshal...Implementation` pattern.

## Generation source

`Azure/azure-rest-api-specs` commit `a3be4b61f3eba3138075131e2669ab43e7841eaf`, at
`specification/cloudhealth/resource-manager/Microsoft.CloudHealth/CloudHealth/stable/2026-11-01/cloudhealth.json`.

Generated with `hashicorp/pandora` commit `ae083a91e9513b6ff31ab09965a31a0d9d70b1dd`.
The six service packages include the generator's resource ID tests without manual changes.

The stable API removes `lookBackWindow`; the provider no longer accepts `look_back_window`.
Dependency aggregation supports `BestOf`, which does not accept thresholds or a unit.
