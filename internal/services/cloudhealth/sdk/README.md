# `Microsoft.CloudHealth` SDK

This directory holds a generated Azure Resource Manager client for `Microsoft.CloudHealth`, API
version `2026-11-01` (stable).

## Why it lives here

`github.com/hashicorp/go-azure-sdk/resource-manager` does not yet include `Microsoft.CloudHealth`.
The provider resources use this local SDK until the upstream package is available.

## Shape

The tree is generated in the exact shape `pandora`'s `generator-go-sdk` emits, so that when a
`go-azure-sdk` package does appear this directory can be deleted and the imports repointed with no
other change:

```
internal/services/cloudhealth/sdk/2026-11-01/<package>/
  →  github.com/hashicorp/go-azure-sdk/resource-manager/cloudhealth/2026-11-01/<package>/
```

Consequences of that, which are deliberate and should be preserved:

- Package files follow the generated layout: `client.go`, `version.go`, `constants.go`, `id_*.go`,
  `method_*.go`, `model_*.go`, `predicates.go`.
- Files carry the Microsoft/MIT header used by generated `go-azure-sdk` code, not the provider's
  own header.
- Nothing here imports `github.com/hashicorp/terraform-provider-azurerm`. The dependency arrow
  points one way only, exactly as it would if this were an external module.
- Discriminated types (`AuthenticationSettingProperties`, `SignalInstanceProperties`,
  `SignalDefinitionProperties`, `DiscoveryRuleSpecification`) use the generated interface +
  `Base…Impl` + `Raw…Impl` + `Unmarshal…Implementation` pattern rather than a hand-rolled union.

## Source of truth

`Azure/azure-rest-api-specs` commit `a3be4b61f3eba3138075131e2669ab43e7841eaf`, at
`specification/cloudhealth/resource-manager/Microsoft.CloudHealth/CloudHealth/stable/2026-11-01/cloudhealth.json`.

Generated with `hashicorp/pandora` commit `ae083a91e9513b6ff31ab09965a31a0d9d70b1dd`.
The six service packages include the generator's resource ID tests without manual changes.

The stable API removes `lookBackWindow`; the provider no longer accepts `look_back_window`.
Dependency aggregation supports `BestOf`, which does not accept thresholds or a unit.
