// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"github.com/hashicorp/terraform-provider-azurerm/internal/common"
)

// Client holds the SDK clients for the `Microsoft.CloudHealth` service.
//
// It is intentionally empty: `Microsoft.CloudHealth` only publishes preview API versions
// (`2025-05-01-preview`, `2026-01-01-preview`, `2026-05-01-preview`), so no `hashicorp/go-azure-sdk`
// package has been generated for it and importing a preview SDK path is rejected by the
// provider's preview API version linter. Once a stable API version ships and the SDK is
// generated, the typed clients are added here and constructed in NewClient below.
type Client struct{}

func NewClient(_ *common.ClientOptions) (*Client, error) {
	return &Client{}, nil
}
