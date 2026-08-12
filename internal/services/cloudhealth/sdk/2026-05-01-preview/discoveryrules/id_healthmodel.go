package discoveryrules

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/recaser"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/resourceids"
)

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See NOTICE.txt in the project root for license information.

func init() {
	recaser.RegisterResourceId(&HealthModelId{})
}

var _ resourceids.ResourceId = &HealthModelId{}

// HealthModelId is a struct representing the Resource ID for a Health Model
type HealthModelId struct {
	SubscriptionId    string
	ResourceGroupName string
	HealthModelName   string
}

// NewHealthModelID returns a new HealthModelId struct
func NewHealthModelID(subscriptionId string, resourceGroupName string, healthModelName string) HealthModelId {
	return HealthModelId{
		SubscriptionId:    subscriptionId,
		ResourceGroupName: resourceGroupName,
		HealthModelName:   healthModelName,
	}
}

// ParseHealthModelID parses 'input' into a HealthModelId
func ParseHealthModelID(input string) (*HealthModelId, error) {
	parser := resourceids.NewParserFromResourceIdType(&HealthModelId{})
	parsed, err := parser.Parse(input, false)
	if err != nil {
		return nil, fmt.Errorf("parsing %q: %+v", input, err)
	}

	id := HealthModelId{}
	if err = id.FromParseResult(*parsed); err != nil {
		return nil, err
	}

	return &id, nil
}

// ParseHealthModelIDInsensitively parses 'input' case-insensitively into a HealthModelId
// note: this method should only be used for API response data and not user input
func ParseHealthModelIDInsensitively(input string) (*HealthModelId, error) {
	parser := resourceids.NewParserFromResourceIdType(&HealthModelId{})
	parsed, err := parser.Parse(input, true)
	if err != nil {
		return nil, fmt.Errorf("parsing %q: %+v", input, err)
	}

	id := HealthModelId{}
	if err = id.FromParseResult(*parsed); err != nil {
		return nil, err
	}

	return &id, nil
}

func (id *HealthModelId) FromParseResult(input resourceids.ParseResult) error {
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

	return nil
}

// ValidateHealthModelID checks that 'input' can be parsed as a Health Model ID
func ValidateHealthModelID(input interface{}, key string) (warnings []string, errors []error) {
	v, ok := input.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected %q to be a string", key))
		return
	}

	if _, err := ParseHealthModelID(v); err != nil {
		errors = append(errors, err)
	}

	return
}

// ID returns the formatted Health Model ID
func (id HealthModelId) ID() string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.CloudHealth/healthmodels/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroupName, id.HealthModelName)
}

// Segments returns a slice of Resource ID Segments which comprise this Health Model ID
func (id HealthModelId) Segments() []resourceids.Segment {
	return []resourceids.Segment{
		resourceids.StaticSegment("staticSubscriptions", "subscriptions", "subscriptions"),
		resourceids.SubscriptionIdSegment("subscriptionId", "12345678-1234-9876-4563-123456789012"),
		resourceids.StaticSegment("staticResourceGroups", "resourceGroups", "resourceGroups"),
		resourceids.ResourceGroupSegment("resourceGroupName", "example-resource-group"),
		resourceids.StaticSegment("staticProviders", "providers", "providers"),
		resourceids.ResourceProviderSegment("staticMicrosoftCloudHealth", "Microsoft.CloudHealth", "Microsoft.CloudHealth"),
		resourceids.StaticSegment("staticHealthmodels", "healthmodels", "healthmodels"),
		resourceids.UserSpecifiedSegment("healthModelName", "healthModelName"),
	}
}

// String returns a human-readable description of this Health Model ID
func (id HealthModelId) String() string {
	components := []string{
		fmt.Sprintf("Subscription: %q", id.SubscriptionId),
		fmt.Sprintf("Resource Group Name: %q", id.ResourceGroupName),
		fmt.Sprintf("Health Model Name: %q", id.HealthModelName),
	}
	return fmt.Sprintf("Health Model (%s)", strings.Join(components, "\n"))
}
