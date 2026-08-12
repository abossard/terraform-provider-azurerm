package authenticationsettings

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/recaser"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/resourceids"
)

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See NOTICE.txt in the project root for license information.

func init() {
	recaser.RegisterResourceId(&AuthenticationSettingId{})
}

var _ resourceids.ResourceId = &AuthenticationSettingId{}

// AuthenticationSettingId is a struct representing the Resource ID for a Authentication Setting
type AuthenticationSettingId struct {
	SubscriptionId            string
	ResourceGroupName         string
	HealthModelName           string
	AuthenticationSettingName string
}

// NewAuthenticationSettingID returns a new AuthenticationSettingId struct
func NewAuthenticationSettingID(subscriptionId string, resourceGroupName string, healthModelName string, authenticationSettingName string) AuthenticationSettingId {
	return AuthenticationSettingId{
		SubscriptionId:            subscriptionId,
		ResourceGroupName:         resourceGroupName,
		HealthModelName:           healthModelName,
		AuthenticationSettingName: authenticationSettingName,
	}
}

// ParseAuthenticationSettingID parses 'input' into a AuthenticationSettingId
func ParseAuthenticationSettingID(input string) (*AuthenticationSettingId, error) {
	parser := resourceids.NewParserFromResourceIdType(&AuthenticationSettingId{})
	parsed, err := parser.Parse(input, false)
	if err != nil {
		return nil, fmt.Errorf("parsing %q: %+v", input, err)
	}

	id := AuthenticationSettingId{}
	if err = id.FromParseResult(*parsed); err != nil {
		return nil, err
	}

	return &id, nil
}

// ParseAuthenticationSettingIDInsensitively parses 'input' case-insensitively into a AuthenticationSettingId
// note: this method should only be used for API response data and not user input
func ParseAuthenticationSettingIDInsensitively(input string) (*AuthenticationSettingId, error) {
	parser := resourceids.NewParserFromResourceIdType(&AuthenticationSettingId{})
	parsed, err := parser.Parse(input, true)
	if err != nil {
		return nil, fmt.Errorf("parsing %q: %+v", input, err)
	}

	id := AuthenticationSettingId{}
	if err = id.FromParseResult(*parsed); err != nil {
		return nil, err
	}

	return &id, nil
}

func (id *AuthenticationSettingId) FromParseResult(input resourceids.ParseResult) error {
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

	if id.AuthenticationSettingName, ok = input.Parsed["authenticationSettingName"]; !ok {
		return resourceids.NewSegmentNotSpecifiedError(id, "authenticationSettingName", input)
	}

	return nil
}

// ValidateAuthenticationSettingID checks that 'input' can be parsed as a Authentication Setting ID
func ValidateAuthenticationSettingID(input interface{}, key string) (warnings []string, errors []error) {
	v, ok := input.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected %q to be a string", key))
		return
	}

	if _, err := ParseAuthenticationSettingID(v); err != nil {
		errors = append(errors, err)
	}

	return
}

// ID returns the formatted Authentication Setting ID
func (id AuthenticationSettingId) ID() string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.CloudHealth/healthmodels/%s/authenticationsettings/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroupName, id.HealthModelName, id.AuthenticationSettingName)
}

// Segments returns a slice of Resource ID Segments which comprise this Authentication Setting ID
func (id AuthenticationSettingId) Segments() []resourceids.Segment {
	return []resourceids.Segment{
		resourceids.StaticSegment("staticSubscriptions", "subscriptions", "subscriptions"),
		resourceids.SubscriptionIdSegment("subscriptionId", "12345678-1234-9876-4563-123456789012"),
		resourceids.StaticSegment("staticResourceGroups", "resourceGroups", "resourceGroups"),
		resourceids.ResourceGroupSegment("resourceGroupName", "example-resource-group"),
		resourceids.StaticSegment("staticProviders", "providers", "providers"),
		resourceids.ResourceProviderSegment("staticMicrosoftCloudHealth", "Microsoft.CloudHealth", "Microsoft.CloudHealth"),
		resourceids.StaticSegment("staticHealthmodels", "healthmodels", "healthmodels"),
		resourceids.UserSpecifiedSegment("healthModelName", "healthModelName"),
		resourceids.StaticSegment("staticAuthenticationsettings", "authenticationsettings", "authenticationsettings"),
		resourceids.UserSpecifiedSegment("authenticationSettingName", "authenticationSettingName"),
	}
}

// String returns a human-readable description of this Authentication Setting ID
func (id AuthenticationSettingId) String() string {
	components := []string{
		fmt.Sprintf("Subscription: %q", id.SubscriptionId),
		fmt.Sprintf("Resource Group Name: %q", id.ResourceGroupName),
		fmt.Sprintf("Health Model Name: %q", id.HealthModelName),
		fmt.Sprintf("Authentication Setting Name: %q", id.AuthenticationSettingName),
	}
	return fmt.Sprintf("Authentication Setting (%s)", strings.Join(components, "\n"))
}
