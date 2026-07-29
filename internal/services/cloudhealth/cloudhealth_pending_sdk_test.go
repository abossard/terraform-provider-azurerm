// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cloudhealth_test

import (
	"errors"
)

// errSDKUnavailable is wrapped by every `Exists` implementation in this package.
//
// `Microsoft.CloudHealth` publishes preview API versions only, so `hashicorp/pandora` has not
// generated a `hashicorp/go-azure-sdk` package for it and there is no typed client to call. The
// acceptance tests in this package are therefore scaffolding: they compile, they parse the resource
// ID through the provider's own parser, and they are skipped unless `TF_ACC` is set. Once a stable
// API version ships and the SDK is vendored, each `Exists` becomes the usual typed `Get` call.
//
// This deliberately fails loudly rather than returning `true`, so a scaffolded run can never be
// mistaken for a real one.
var errSDKUnavailable = errors.New("the Microsoft.CloudHealth SDK is not yet available: no stable API version has shipped, so hashicorp/go-azure-sdk has no cloudhealth package to call")
