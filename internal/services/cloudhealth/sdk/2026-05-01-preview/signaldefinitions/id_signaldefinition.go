package signaldefinitions

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/recaser"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/resourceids"
)

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See NOTICE.txt in the project root for license information.

func init() {
	recaser.RegisterResourceId(&SignalDefinitionId{})
}

var _ resourceids.ResourceId = &SignalDefinitionId{}

// SignalDefinitionId is a struct representing the Resource ID for a Signal Definition
type SignalDefinitionId struct {
	SubscriptionId       string
	ResourceGroupName    string
	HealthModelName      string
	SignalDefinitionName string
}

// NewSignalDefinitionID returns a new SignalDefinitionId struct
func NewSignalDefinitionID(subscriptionId string, resourceGroupName string, healthModelName string, signalDefinitionName string) SignalDefinitionId {
	return SignalDefinitionId{
		SubscriptionId:       subscriptionId,
		ResourceGroupName:    resourceGroupName,
		HealthModelName:      healthModelName,
		SignalDefinitionName: signalDefinitionName,
	}
}

// ParseSignalDefinitionID parses 'input' into a SignalDefinitionId
func ParseSignalDefinitionID(input string) (*SignalDefinitionId, error) {
	parser := resourceids.NewParserFromResourceIdType(&SignalDefinitionId{})
	parsed, err := parser.Parse(input, false)
	if err != nil {
		return nil, fmt.Errorf("parsing %q: %+v", input, err)
	}

	id := SignalDefinitionId{}
	if err = id.FromParseResult(*parsed); err != nil {
		return nil, err
	}

	return &id, nil
}

// ParseSignalDefinitionIDInsensitively parses 'input' case-insensitively into a SignalDefinitionId
// note: this method should only be used for API response data and not user input
func ParseSignalDefinitionIDInsensitively(input string) (*SignalDefinitionId, error) {
	parser := resourceids.NewParserFromResourceIdType(&SignalDefinitionId{})
	parsed, err := parser.Parse(input, true)
	if err != nil {
		return nil, fmt.Errorf("parsing %q: %+v", input, err)
	}

	id := SignalDefinitionId{}
	if err = id.FromParseResult(*parsed); err != nil {
		return nil, err
	}

	return &id, nil
}

func (id *SignalDefinitionId) FromParseResult(input resourceids.ParseResult) error {
	var ok bool

	if id.SubscriptionId, ok = input.Parsed["subscriptionId"]; !ok {
		return resourceids.NewSegmentNotSpecifiedError(id, "subscriptionId", input)
	}

	if id.ResourceGroupName, ok = input.Parsed["resourceGroupName"]; !ok {
		return resourceids.NewSegmentNotSpecifiedError(id, "resourceGroupName", input)
	}

	if id.HealthModelName, ok = input.Parsed["healthModelName"]; !ok {
		return resourceids.NewSegmentNotSpecifiedError(id, "healthModelName", input)
	}

	if id.SignalDefinitionName, ok = input.Parsed["signalDefinitionName"]; !ok {
		return resourceids.NewSegmentNotSpecifiedError(id, "signalDefinitionName", input)
	}

	return nil
}

// ValidateSignalDefinitionID checks that 'input' can be parsed as a Signal Definition ID
func ValidateSignalDefinitionID(input interface{}, key string) (warnings []string, errors []error) {
	v, ok := input.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected %q to be a string", key))
		return
	}

	if _, err := ParseSignalDefinitionID(v); err != nil {
		errors = append(errors, err)
	}

	return
}

// ID returns the formatted Signal Definition ID
func (id SignalDefinitionId) ID() string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.CloudHealth/healthmodels/%s/signaldefinitions/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroupName, id.HealthModelName, id.SignalDefinitionName)
}

// Segments returns a slice of Resource ID Segments which comprise this Signal Definition ID
func (id SignalDefinitionId) Segments() []resourceids.Segment {
	return []resourceids.Segment{
		resourceids.StaticSegment("staticSubscriptions", "subscriptions", "subscriptions"),
		resourceids.SubscriptionIdSegment("subscriptionId", "12345678-1234-9876-4563-123456789012"),
		resourceids.StaticSegment("staticResourceGroups", "resourceGroups", "resourceGroups"),
		resourceids.ResourceGroupSegment("resourceGroupName", "example-resource-group"),
		resourceids.StaticSegment("staticProviders", "providers", "providers"),
		resourceids.ResourceProviderSegment("staticMicrosoftCloudHealth", "Microsoft.CloudHealth", "Microsoft.CloudHealth"),
		resourceids.StaticSegment("staticHealthmodels", "healthmodels", "healthmodels"),
		resourceids.UserSpecifiedSegment("healthModelName", "healthModelName"),
		resourceids.StaticSegment("staticSignaldefinitions", "signaldefinitions", "signaldefinitions"),
		resourceids.UserSpecifiedSegment("signalDefinitionName", "signalDefinitionName"),
	}
}

// String returns a human-readable description of this Signal Definition ID
func (id SignalDefinitionId) String() string {
	components := []string{
		fmt.Sprintf("Subscription: %q", id.SubscriptionId),
		fmt.Sprintf("Resource Group Name: %q", id.ResourceGroupName),
		fmt.Sprintf("Health Model Name: %q", id.HealthModelName),
		fmt.Sprintf("Signal Definition Name: %q", id.SignalDefinitionName),
	}
	return fmt.Sprintf("Signal Definition (%s)", strings.Join(components, "\n"))
}
