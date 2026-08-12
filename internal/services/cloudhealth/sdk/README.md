# `Microsoft.CloudHealth` SDK

This directory holds a generated Azure Resource Manager client for `Microsoft.CloudHealth`, API
version `2026-05-01-preview`.

## Why it lives here

`Microsoft.CloudHealth` only publishes preview API versions, so `hashicorp/pandora` has not yet
emitted a package for it under `github.com/hashicorp/go-azure-sdk/resource-manager`. The provider
resources in `internal/services/cloudhealth` need a client today, so one is vendored in-repo.

## Shape

The tree is generated in the exact shape `pandora`'s `generator-go-sdk` emits, so that when a
`go-azure-sdk` package does appear this directory can be deleted and the imports repointed with no
other change:

```
internal/services/cloudhealth/sdk/2026-05-01-preview/<package>/
  →  github.com/hashicorp/go-azure-sdk/resource-manager/cloudhealth/2026-05-01-preview/<package>/
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

`Azure/azure-rest-api-specs`, at
`specification/cloudhealth/resource-manager/Microsoft.CloudHealth/CloudHealth/preview/2026-05-01-preview/cloudhealth.json`.
