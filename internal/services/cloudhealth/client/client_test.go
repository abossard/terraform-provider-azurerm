// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"net/http"
	"testing"
	"time"

	sdkclient "github.com/hashicorp/go-azure-sdk/sdk/client"
	"github.com/hashicorp/go-azure-sdk/sdk/client/resourcemanager"
	"github.com/hashicorp/go-azure-sdk/sdk/environments"
	"github.com/hashicorp/terraform-provider-azurerm/internal/common"
)

func TestClientStableAPIVersion(t *testing.T) {
	client, err := NewClient(&common.ClientOptions{
		Authorizers: &common.Authorizers{},
		Environment: *environments.AzurePublic(),
	})
	if err != nil {
		t.Fatalf("building CloudHealth clients: %+v", err)
	}

	testData := []struct {
		name   string
		client *resourcemanager.Client
	}{
		{name: "authentication settings", client: client.AuthenticationSettings.Client},
		{name: "discovery rules", client: client.DiscoveryRules.Client},
		{name: "entities", client: client.Entities.Client},
		{name: "health models", client: client.HealthModels.Client},
		{name: "relationships", client: client.Relationships.Client},
		{name: "signal definitions", client: client.SignalDefinitions.Client},
	}

	for _, v := range testData {
		t.Run(v.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()

			request, err := v.client.NewRequest(ctx, sdkclient.RequestOptions{
				ContentType:         "application/json",
				ExpectedStatusCodes: []int{http.StatusOK},
				HttpMethod:          http.MethodGet,
				Path:                "/",
			})
			if err != nil {
				t.Fatalf("building request: %+v", err)
			}

			if actual := request.URL.Query().Get("api-version"); actual != "2026-11-01" {
				t.Fatalf("expected stable API version 2026-11-01, got %q", actual)
			}
		})
	}
}
