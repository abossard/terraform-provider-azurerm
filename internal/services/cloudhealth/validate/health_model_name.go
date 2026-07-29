// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package validate

import (
	"fmt"
	"regexp"
)

// healthModelNamePattern is the `healthModelName` path parameter pattern from the
// `2026-05-01-preview` Microsoft.CloudHealth swagger.
var healthModelNamePattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]{1,42}[a-zA-Z0-9]$`)

func HealthModelName(v interface{}, k string) (warnings []string, errors []error) {
	value, ok := v.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected %q to be a string", k))
		return warnings, errors
	}

	if !healthModelNamePattern.MatchString(value) {
		errors = append(errors, fmt.Errorf("%q must be between 3 and 44 characters long, start with a letter, end with a letter or number, and may only contain letters, numbers and hyphens, got %q", k, value))
	}

	return warnings, errors
}
