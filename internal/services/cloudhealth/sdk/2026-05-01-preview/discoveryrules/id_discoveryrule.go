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
	recaser.RegisterResourceId(&DiscoveryRuleId{})
}

var _ resourceids.ResourceId = &DiscoveryRuleId{}

// DiscoveryRuleId is a struct representing the Resource ID for a Discovery Rule
type DiscoveryRuleId struct {
	SubscriptionId    string
	ResourceGroupName string
	HealthModelName   string
	DiscoveryRuleName string
}

// NewDiscoveryRuleID returns a new DiscoveryRuleId struct
func NewDiscoveryRuleID(subscriptionId string, resourceGroupName string, healthModelName string, discoveryRuleName string) DiscoveryRuleId {
	return DiscoveryRuleId{
		SubscriptionId:    subscriptionId,
		ResourceGroupName: resourceGroupName,
		HealthModelName:   healthModelName,
		DiscoveryRuleName: discoveryRuleName,
	}
}

// ParseDiscoveryRuleID parses 'input' into a DiscoveryRuleId
func ParseDiscoveryRuleID(input string) (*DiscoveryRuleId, error) {
	parser := resourceids.NewParserFromResourceIdType(&DiscoveryRuleId{})
	parsed, err := parser.Parse(input, false)
	if err != nil {
		return nil, fmt.Errorf("parsing %q: %+v", input, err)
	}

	id := DiscoveryRuleId{}
	if err = id.FromParseResult(*parsed); err != nil {
		return nil, err
	}

	return &id, nil
}

// ParseDiscoveryRuleIDInsensitively parses 'input' case-insensitively into a DiscoveryRuleId
// note: this method should only be used for API response data and not user input
func ParseDiscoveryRuleIDInsensitively(input string) (*DiscoveryRuleId, error) {
	parser := resourceids.NewParserFromResourceIdType(&DiscoveryRuleId{})
	parsed, err := parser.Parse(input, true)
	if err != nil {
		return nil, fmt.Errorf("parsing %q: %+v", input, err)
	}

	id := DiscoveryRuleId{}
	if err = id.FromParseResult(*parsed); err != nil {
		return nil, err
	}

	return &id, nil
}

func (id *DiscoveryRuleId) FromParseResult(input resourceids.ParseResult) error {
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

	if id.DiscoveryRuleName, ok = input.Parsed["discoveryRuleName"]; !ok {
		return resourceids.NewSegmentNotSpecifiedError(id, "discoveryRuleName", input)
	}

	return nil
}

// ValidateDiscoveryRuleID checks that 'input' can be parsed as a Discovery Rule ID
func ValidateDiscoveryRuleID(input interface{}, key string) (warnings []string, errors []error) {
	v, ok := input.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected %q to be a string", key))
		return
	}

	if _, err := ParseDiscoveryRuleID(v); err != nil {
		errors = append(errors, err)
	}

	return
}

// ID returns the formatted Discovery Rule ID
func (id DiscoveryRuleId) ID() string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.CloudHealth/healthmodels/%s/discoveryrules/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroupName, id.HealthModelName, id.DiscoveryRuleName)
}

// Segments returns a slice of Resource ID Segments which comprise this Discovery Rule ID
func (id DiscoveryRuleId) Segments() []resourceids.Segment {
	return []resourceids.Segment{
		resourceids.StaticSegment("staticSubscriptions", "subscriptions", "subscriptions"),
		resourceids.SubscriptionIdSegment("subscriptionId", "12345678-1234-9876-4563-123456789012"),
		resourceids.StaticSegment("staticResourceGroups", "resourceGroups", "resourceGroups"),
		resourceids.ResourceGroupSegment("resourceGroupName", "example-resource-group"),
		resourceids.StaticSegment("staticProviders", "providers", "providers"),
		resourceids.ResourceProviderSegment("staticMicrosoftCloudHealth", "Microsoft.CloudHealth", "Microsoft.CloudHealth"),
		resourceids.StaticSegment("staticHealthmodels", "healthmodels", "healthmodels"),
		resourceids.UserSpecifiedSegment("healthModelName", "healthModelName"),
		resourceids.StaticSegment("staticDiscoveryrules", "discoveryrules", "discoveryrules"),
		resourceids.UserSpecifiedSegment("discoveryRuleName", "discoveryRuleName"),
	}
}

// String returns a human-readable description of this Discovery Rule ID
func (id DiscoveryRuleId) String() string {
	components := []string{
		fmt.Sprintf("Subscription: %q", id.SubscriptionId),
		fmt.Sprintf("Resource Group Name: %q", id.ResourceGroupName),
		fmt.Sprintf("Health Model Name: %q", id.HealthModelName),
		fmt.Sprintf("Discovery Rule Name: %q", id.DiscoveryRuleName),
	}
	return fmt.Sprintf("Discovery Rule (%s)", strings.Join(components, "\n"))
}
