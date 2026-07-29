// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package parse

// NOTE: this file is generated via 'go:generate' - manual changes will be overwritten

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/resourceids"
)

type SignalDefinitionId struct {
	SubscriptionId  string
	ResourceGroup   string
	HealthmodelName string
	Name            string
}

func NewSignalDefinitionID(subscriptionId, resourceGroup, healthmodelName, name string) SignalDefinitionId {
	return SignalDefinitionId{
		SubscriptionId:  subscriptionId,
		ResourceGroup:   resourceGroup,
		HealthmodelName: healthmodelName,
		Name:            name,
	}
}

func (id SignalDefinitionId) String() string {
	segments := []string{
		fmt.Sprintf("Name %q", id.Name),
		fmt.Sprintf("Healthmodel Name %q", id.HealthmodelName),
		fmt.Sprintf("Resource Group %q", id.ResourceGroup),
	}
	segmentsStr := strings.Join(segments, " / ")
	return fmt.Sprintf("%s: (%s)", "Signal Definition", segmentsStr)
}

func (id SignalDefinitionId) ID() string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.CloudHealth/healthmodels/%s/signaldefinitions/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroup, id.HealthmodelName, id.Name)
}

// SignalDefinitionID parses a SignalDefinition ID into an SignalDefinitionId struct
func SignalDefinitionID(input string) (*SignalDefinitionId, error) {
	id, err := resourceids.ParseAzureResourceID(input)
	if err != nil {
		return nil, fmt.Errorf("parsing %q as an SignalDefinition ID: %+v", input, err)
	}

	resourceId := SignalDefinitionId{
		SubscriptionId: id.SubscriptionID,
		ResourceGroup:  id.ResourceGroup,
	}

	if resourceId.SubscriptionId == "" {
		return nil, errors.New("ID was missing the 'subscriptions' element")
	}

	if resourceId.ResourceGroup == "" {
		return nil, errors.New("ID was missing the 'resourceGroups' element")
	}

	if resourceId.HealthmodelName, err = id.PopSegment("healthmodels"); err != nil {
		return nil, err
	}
	if resourceId.Name, err = id.PopSegment("signaldefinitions"); err != nil {
		return nil, err
	}

	if err := id.ValidateNoEmptySegments(input); err != nil {
		return nil, err
	}

	return &resourceId, nil
}
