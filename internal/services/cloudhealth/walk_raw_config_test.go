// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cloudhealth

import (
	"strings"
	"testing"

	"github.com/hashicorp/go-cty/cty"
)

// TestWalkRawConfigBlocks exercises the descent itself rather than either guard's rules. Both
// resource-level entry points reach their validators only through this walk, and the blocks they
// guard sit at several different depths and inside different collection types, so a walk that
// stopped early would disable a guard silently while every direct validator test still passed.
func TestWalkRawConfigBlocks(t *testing.T) {
	// The entity schema nests `evaluation_rule` inside a list of `signal_group` blocks, inside a
	// list of `signal` blocks. `dependency` sits one level shallower, directly under `signal_group`.
	entityConfig := cty.ObjectVal(map[string]cty.Value{
		"name": cty.StringVal("example"),
		"signal_group": cty.ListVal([]cty.Value{
			cty.ObjectVal(map[string]cty.Value{
				"dependency": cty.ListVal([]cty.Value{
					cty.ObjectVal(withNullAttributes(map[string]cty.Value{
						"aggregation_type": cty.StringVal("WorstOf"),
					})),
				}),
				"azure_resource": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"signal": cty.ListVal([]cty.Value{
							cty.ObjectVal(map[string]cty.Value{
								"signal_definition_name": cty.StringVal("cpu"),
								"evaluation_rule":        cty.NullVal(cty.DynamicPseudoType),
							}),
						}),
					}),
				}),
			}),
		}),
	})

	testData := []struct {
		name          string
		input         cty.Value
		blockName     string
		expectedCalls int
	}{
		{
			name:          "a block nested two lists deep is reached",
			input:         entityConfig,
			blockName:     "signal",
			expectedCalls: 1,
		},
		{
			name:          "a block nested one list deep is reached",
			input:         entityConfig,
			blockName:     "dependency",
			expectedCalls: 1,
		},
		{
			name:          "a block name that does not occur is never dispatched",
			input:         entityConfig,
			blockName:     "alert",
			expectedCalls: 0,
		},
		{
			name:          "a null configuration is tolerated",
			input:         cty.NullVal(cty.EmptyObject),
			blockName:     "dependency",
			expectedCalls: 0,
		},
		{
			name:          "an unknown configuration is tolerated",
			input:         cty.UnknownVal(cty.EmptyObject),
			blockName:     "dependency",
			expectedCalls: 0,
		},
		{
			name:          "an empty collection is tolerated",
			input:         cty.ObjectVal(map[string]cty.Value{"signal_group": cty.ListValEmpty(cty.EmptyObject)}),
			blockName:     "dependency",
			expectedCalls: 0,
		},
		{
			// `tags` is a `cty.Map(cty.String)` whose keys the user chooses freely, so a tag can
			// carry the same name as a block. Dispatching on the name alone would hand the
			// validator a string and crash the provider.
			name: "a tag sharing a block's name is not dispatched",
			input: cty.ObjectVal(map[string]cty.Value{
				"name": cty.StringVal("example"),
				"tags": cty.MapVal(map[string]cty.Value{
					"env":        cty.StringVal("prod"),
					"dependency": cty.StringVal("payments-api"),
				}),
			}),
			blockName:     "dependency",
			expectedCalls: 0,
		},
		{
			name: "a scalar attribute sharing a block's name is not dispatched",
			input: cty.ObjectVal(map[string]cty.Value{
				"dependency": cty.StringVal("payments-api"),
			}),
			blockName:     "dependency",
			expectedCalls: 0,
		},
		{
			name: "a list of strings sharing a block's name is not dispatched",
			input: cty.ObjectVal(map[string]cty.Value{
				"dependency": cty.ListVal([]cty.Value{cty.StringVal("payments-api")}),
			}),
			blockName:     "dependency",
			expectedCalls: 0,
		},
		{
			name: "a block is still reached when a tag shares another block's name",
			input: cty.ObjectVal(map[string]cty.Value{
				"tags": cty.MapVal(map[string]cty.Value{"dependency": cty.StringVal("payments-api")}),
				"signal_group": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"dependency": cty.ListVal([]cty.Value{
							cty.ObjectVal(withNullAttributes(map[string]cty.Value{
								"aggregation_type": cty.StringVal("WorstOf"),
							})),
						}),
					}),
				}),
			}),
			blockName:     "dependency",
			expectedCalls: 1,
		},
	}

	for _, v := range testData {
		t.Run(v.name, func(t *testing.T) {
			calls := 0
			err := walkRawConfigBlocks(v.input, map[string]func(cty.Value) error{
				v.blockName: func(cty.Value) error {
					calls++
					return nil
				},
			})
			if err != nil {
				t.Fatalf("expected no error but got: %+v", err)
			}

			if calls != v.expectedCalls {
				t.Fatalf("expected the validator for %q to be called %d time(s) but it was called %d time(s)", v.blockName, v.expectedCalls, calls)
			}
		})
	}

	t.Run("a validator error stops the walk and is returned", func(t *testing.T) {
		err := walkRawConfigBlocks(entityConfig, map[string]func(cty.Value) error{
			"dependency": func(cty.Value) error {
				return errFromWalkTest
			},
		})
		if err == nil {
			t.Fatal("expected the validator's error to be returned but got nil")
		}

		if !strings.Contains(err.Error(), errFromWalkTest.Error()) {
			t.Fatalf("expected the validator's error but got: %+v", err)
		}
	})

	t.Run("both guards dispatch from a single walk", func(t *testing.T) {
		seen := map[string]int{}
		err := walkRawConfigBlocks(entityConfig, map[string]func(cty.Value) error{
			"dependency": func(cty.Value) error { seen["dependency"]++; return nil },
			"signal":     func(cty.Value) error { seen["signal"]++; return nil },
		})
		if err != nil {
			t.Fatalf("expected no error but got: %+v", err)
		}

		if seen["dependency"] != 1 || seen["signal"] != 1 {
			t.Fatalf("expected each validator to be called once but got %+v", seen)
		}
	})

	t.Run("the real validators do not crash on a tag that shares their block's name", func(t *testing.T) {
		// A counting validator would pass even if the walk dispatched a string, because it never
		// looks at the value. The shipped validators do, and a `cty.String` reaching one of them
		// panics the provider mid-plan, so they are used here verbatim.
		input := cty.ObjectVal(map[string]cty.Value{
			"name": cty.StringVal("example"),
			"tags": cty.MapVal(map[string]cty.Value{
				"env":             cty.StringVal("prod"),
				"dependency":      cty.StringVal("payments-api"),
				"evaluation_rule": cty.StringVal("still-a-tag"),
			}),
		})

		if err := validateDependencyRawConfig(input); err != nil {
			t.Fatalf("expected no error but got: %+v", err)
		}

		if err := validateEvaluationRuleRawConfig(input); err != nil {
			t.Fatalf("expected no error but got: %+v", err)
		}

		if dependencyDegradedThresholdConfigured(input) {
			t.Fatal("expected a tag named `dependency` not to be read as a configured threshold")
		}
	})
}

var errFromWalkTest = &walkTestError{}

type walkTestError struct{}

func (*walkTestError) Error() string {
	return "validator refused the block"
}
