// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package validate

import (
	"strings"
	"testing"
)

func TestHealthModelName(t *testing.T) {
	cases := []struct {
		Name     string
		Value    string
		ErrCount int
	}{
		{Name: "minimum length", Value: "abc", ErrCount: 0},
		{Name: "letters numbers and hyphens", Value: "hm-workload-01", ErrCount: 0},
		{Name: "maximum length", Value: "a" + strings.Repeat("b", 42) + "c", ErrCount: 0},
		{Name: "ends with a number", Value: "hm1", ErrCount: 0},
		{Name: "empty", Value: "", ErrCount: 1},
		{Name: "too short", Value: "ab", ErrCount: 1},
		{Name: "too long", Value: "a" + strings.Repeat("b", 43) + "c", ErrCount: 1},
		{Name: "starts with a number", Value: "1hm", ErrCount: 1},
		{Name: "starts with a hyphen", Value: "-hm", ErrCount: 1},
		{Name: "ends with a hyphen", Value: "hm-", ErrCount: 1},
		{Name: "contains an underscore", Value: "hm_workload", ErrCount: 1},
		{Name: "contains a period", Value: "hm.workload", ErrCount: 1},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			_, errors := HealthModelName(tc.Value, "name")
			if len(errors) != tc.ErrCount {
				t.Fatalf("expected HealthModelName(%q) to return %d errors, got %d: %v", tc.Value, tc.ErrCount, len(errors), errors)
			}
		})
	}
}

func TestHealthModelNameRejectsNonString(t *testing.T) {
	_, errors := HealthModelName(42, "name")
	if len(errors) != 1 {
		t.Fatalf("expected HealthModelName(42) to return 1 error, got %d: %v", len(errors), errors)
	}
}
