// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package validate

import (
	"fmt"
	"regexp"
)

// healthModelChildNamePattern is declared identically by the `authenticationSettingName`,
// `discoveryRuleName`, `entityName`, `signalDefinitionName` and `signalName` definitions and by the
// `relationshipName` path parameter in the `2026-05-01-preview` Microsoft.CloudHealth swagger. All six
// carry the same pattern, so one validator covers every child resource name.
var healthModelChildNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{1,258}[a-zA-Z0-9]$`)

func HealthModelChildName(v interface{}, k string) (warnings []string, errors []error) {
	value, ok := v.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected %q to be a string", k))
		return warnings, errors
	}

	if !healthModelChildNamePattern.MatchString(value) {
		errors = append(errors, fmt.Errorf("%q must be between 3 and 260 characters long, start and end with a letter or number, and may only contain letters, numbers and hyphens, got %q", k, value))
	}

	return warnings, errors
}
