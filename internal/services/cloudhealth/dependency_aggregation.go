// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cloudhealth

import (
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/cloudhealth/sdk/2026-11-01/entities"
)

// validateDependencyRawConfig rejects the `dependency` combinations the service refuses, so the
// operator sees the problem at plan time instead of after a round trip.
//
// Unlike the `evaluation_rule` guard, this one is a usability improvement rather than a correctness
// fix: the service already fails these payloads loudly and names the offending property, so nothing
// is silently dropped and no perpetual diff is possible. It exists because a plan-time refusal is
// cheaper than a failed apply, not because the API is permissive here.
//
// The two constraints are complementary:
//
//	WorstOf / BestOf            -> `unit`, `unhealthy_threshold` and `degraded_threshold` are refused
//	MinHealthy / MaxNotHealthy -> `unit` and `unhealthy_threshold` are required
//
// `degraded_threshold` stays optional in every mode; the swagger describes it as the absence of a
// degraded state rather than a required companion of `unhealthy_threshold`.
func validateDependencyRawConfig(input cty.Value) error {
	return walkRawConfigBlocks(input, map[string]func(cty.Value) error{
		"dependency": validateDependencyBlock,
	})
}

func validateDependencyBlock(input cty.Value) error {
	if input.IsNull() || !input.IsKnown() || input.LengthInt() == 0 {
		return nil
	}

	for _, dependency := range input.AsValueSlice() {
		if err := validateDependencyAggregation(dependency); err != nil {
			return err
		}
	}

	return nil
}

func validateDependencyAggregation(input cty.Value) error {
	if input.IsNull() || !input.IsKnown() || !input.Type().IsObjectType() {
		return nil
	}

	aggregationType := input.GetAttr("aggregation_type")
	if aggregationType.IsNull() || !aggregationType.IsKnown() {
		return nil
	}

	// An unknown value is still a value the user asked to send, so only an explicit null counts as
	// "not configured". The threshold fields are `TypeFloat`, so this distinction survives only in
	// the raw configuration - a decoded `0` could mean either.
	isConfigured := func(name string) bool {
		if !input.Type().HasAttribute(name) {
			return false
		}
		return !input.GetAttr(name).IsNull()
	}

	if aggregationType.AsString() == string(entities.AggregationTypeWorstOf) || aggregationType.AsString() == string(entities.AggregationTypeBestOf) {
		for _, name := range []string{"unit", "unhealthy_threshold", "degraded_threshold"} {
			if isConfigured(name) {
				return fmt.Errorf("`%s` cannot be set in `signal_group.dependency` when `aggregation_type` is `%s` - thresholds and units only apply to the counting aggregations `%s` and `%s`", name, aggregationType.AsString(), entities.AggregationTypeMinHealthy, entities.AggregationTypeMaxNotHealthy)
			}
		}

		return nil
	}

	for _, name := range []string{"unit", "unhealthy_threshold"} {
		if !isConfigured(name) {
			return fmt.Errorf("`%s` must be set in `signal_group.dependency` when `aggregation_type` is `%s` - the service cannot count dependencies without it", name, aggregationType.AsString())
		}
	}

	return nil
}

// dependencyDegradedThresholdConfigured reports whether `signal_group.dependency.degraded_threshold`
// appears in the configuration.
//
// It cannot be answered from the decoded model. The typed decoder skips a *top-level* argument that
// `GetOkExists` reports as absent, so a top-level `*float64` is nil exactly when the argument is
// unset - which is why `health_objective_percentage` can be a pointer. Inside a nested block there
// is no such per-field signal: the decoder walks the flattened block map and sets every field it
// finds, so an absent `TypeFloat` arrives as a pointer to `0` rather than as nil. Making this field
// a pointer therefore produced `"degradedThreshold": 0` on every dependency block, which the service
// rejects outright under `WorstOf`.
//
// The raw configuration is the only place the distinction survives, and the same walk the plan-time
// guards use finds it, so the block is located the same way here.
//
// The answer is one boolean for the whole resource, which is only sound because `signal_group` and
// `dependency` are both `MaxItems: 1`: at most one `dependency` block can exist, so "some block set
// it" and "this block set it" cannot diverge. If either block ever admits more than one item this
// collapses - one configured `degraded_threshold` would switch it on for every sibling - and the
// answer would have to be per-block. `TestDependencyBlocksAreSingletons` fails if that changes.
func dependencyDegradedThresholdConfigured(rawConfig cty.Value) bool {
	configured := false

	// walkRawConfigBlocks only returns an error when a validator does, and this one never fails.
	_ = walkRawConfigBlocks(rawConfig, map[string]func(cty.Value) error{
		"dependency": func(input cty.Value) error {
			if input.IsNull() || !input.IsKnown() || input.LengthInt() == 0 {
				return nil
			}

			for _, dependency := range input.AsValueSlice() {
				if dependency.IsNull() || !dependency.IsKnown() || !dependency.Type().IsObjectType() {
					continue
				}

				if !dependency.Type().HasAttribute("degraded_threshold") {
					continue
				}

				if !dependency.GetAttr("degraded_threshold").IsNull() {
					configured = true
				}
			}

			return nil
		},
	})

	return configured
}
