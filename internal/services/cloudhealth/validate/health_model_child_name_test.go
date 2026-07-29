// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package validate

import (
	"strings"
	"testing"
)

func TestHealthModelChildName(t *testing.T) {
	cases := []struct {
		Name     string
		Value    string
		ErrCount int
	}{
		{Name: "minimum length", Value: "abc", ErrCount: 0},
		{Name: "letters numbers and hyphens", Value: "r-frontend-backend", ErrCount: 0},
		{Name: "starts with a number", Value: "1st-entity", ErrCount: 0},
		{Name: "maximum length", Value: "a" + strings.Repeat("b", 258) + "c", ErrCount: 0},
		{Name: "empty", Value: "", ErrCount: 1},
		{Name: "too short", Value: "ab", ErrCount: 1},
		{Name: "too long", Value: "a" + strings.Repeat("b", 259) + "c", ErrCount: 1},
		{Name: "starts with a hyphen", Value: "-entity", ErrCount: 1},
		{Name: "ends with a hyphen", Value: "entity-", ErrCount: 1},
		{Name: "contains an underscore", Value: "my_entity", ErrCount: 1},
		{Name: "contains a slash", Value: "my/entity", ErrCount: 1},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			_, errors := HealthModelChildName(tc.Value, "name")
			if len(errors) != tc.ErrCount {
				t.Fatalf("expected HealthModelChildName(%q) to return %d errors, got %d: %v", tc.Value, tc.ErrCount, len(errors), errors)
			}
		})
	}
}

func TestHealthModelChildNameRejectsNonString(t *testing.T) {
	_, errors := HealthModelChildName(42, "name")
	if len(errors) != 1 {
		t.Fatalf("expected HealthModelChildName(42) to return 1 error, got %d: %v", len(errors), errors)
	}
}
