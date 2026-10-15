// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cloudhealth

import (
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/cloudhealth/sdk/2026-11-01/entities"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/validation"
)

// EvaluationRuleModel maps the `evaluation_rule` block shared by every signal kind and by
// `azurerm_monitor_health_model_signal_definition`.
type EvaluationRuleModel struct {
	Unhealthy []ThresholdRuleModel `tfschema:"unhealthy"`
	Degraded  []ThresholdRuleModel `tfschema:"degraded"`
}

type ThresholdRuleModel struct {
	Operator    string  `tfschema:"operator"`
	Threshold   float64 `tfschema:"threshold"`
	Sensitivity string  `tfschema:"sensitivity"`
}

func evaluationRuleSchema() *pluginsdk.Schema {
	return &pluginsdk.Schema{
		Type:     pluginsdk.TypeList,
		Required: true,
		MaxItems: 1,
		Elem: &pluginsdk.Resource{
			Schema: map[string]*pluginsdk.Schema{
				"unhealthy": thresholdRuleSchema(true),
				"degraded":  thresholdRuleSchema(false),
			},
		},
	}
}

func instanceEvaluationRuleSchema() *pluginsdk.Schema {
	schema := evaluationRuleSchema()
	schema.Required, schema.Optional = false, true
	return schema
}

func thresholdRuleSchema(required bool) *pluginsdk.Schema {
	return &pluginsdk.Schema{
		Type:     pluginsdk.TypeList,
		Required: required,
		Optional: !required,
		MaxItems: 1,
		Elem: &pluginsdk.Resource{
			Schema: map[string]*pluginsdk.Schema{
				"operator": {
					Type:         pluginsdk.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice(entities.PossibleValuesForSignalOperator(), false),
				},

				"threshold": {
					Type:     pluginsdk.TypeFloat,
					Optional: true,
				},

				"sensitivity": {
					Type:         pluginsdk.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringInSlice(entities.PossibleValuesForDynamicThresholdSensitivity(), false),
				},
			},
		},
	}
}

func expandEvaluationRule(input []EvaluationRuleModel) *entities.EvaluationRule {
	if len(input) == 0 {
		return nil
	}

	rules := input[0]
	output := entities.EvaluationRule{}

	if len(rules.Unhealthy) > 0 {
		output.UnhealthyRule = expandThresholdRule(rules.Unhealthy[0])
	}

	if len(rules.Degraded) > 0 {
		degraded := expandThresholdRule(rules.Degraded[0])
		output.DegradedRule = &degraded
	}

	return &output
}

func expandThresholdRule(input ThresholdRuleModel) entities.ThresholdRuleV2 {
	output := entities.ThresholdRuleV2{
		Operator: entities.SignalOperator(input.Operator),
	}

	// Everything the user configured is forwarded, whether or not it suits the operator. The service
	// rejects each mismatch by name (`DynamicThresholdNotAllowed` for a `threshold` alongside
	// `Dynamic`, and an `ObjectInvalidProperty` naming `sensitivity` alongside any
	// other operator), so forwarding turns a mismatch into a loud apply-time failure. Dropping it
	// instead would discard configuration the Read side cannot restore, leaving a permanent diff.
	// CustomizeDiff still rejects both combinations at plan time whenever `operator` is known; this
	// is the backstop for the case where it is only resolved during apply.
	if input.Sensitivity != "" {
		sensitivity := entities.DynamicThresholdSensitivity(input.Sensitivity)
		output.Sensitivity = &sensitivity
	}

	// `threshold` is a `TypeFloat`, so once decoded an explicit `0` and an absent value are the same
	// value. A zero is therefore only forwarded when the operator uses it; CustomizeDiff already
	// rejects an explicit `threshold = 0` alongside `Dynamic`, so a zero arriving here with `Dynamic`
	// is genuinely absent rather than configured.
	if output.Operator != entities.SignalOperatorDynamic || input.Threshold != 0 {
		threshold := input.Threshold
		output.Threshold = &threshold
	}

	return output
}

func flattenEvaluationRule(input *entities.EvaluationRule) []EvaluationRuleModel {
	if input == nil {
		return []EvaluationRuleModel{}
	}

	output := EvaluationRuleModel{
		Unhealthy: []ThresholdRuleModel{flattenThresholdRule(input.UnhealthyRule)},
	}

	if input.DegradedRule != nil {
		output.Degraded = []ThresholdRuleModel{flattenThresholdRule(*input.DegradedRule)}
	}

	return []EvaluationRuleModel{output}
}

func flattenThresholdRule(input entities.ThresholdRuleV2) ThresholdRuleModel {
	output := ThresholdRuleModel{
		Operator: string(input.Operator),
	}

	if input.Threshold != nil {
		output.Threshold = *input.Threshold
	}

	if input.Sensitivity != nil {
		output.Sensitivity = string(*input.Sensitivity)
	}

	return output
}

// walkRawConfigBlocks descends the raw configuration and hands every nested block whose name is a
// key of `validators` to the matching validator. The blocks these guards care about sit inside list
// blocks several levels down, where the schema-level helpers (`ConflictsWith`, `RequiredWith`,
// `ExactlyOneOf`) cannot reach, so the configuration is walked instead.
//
// The raw configuration is used rather than the decoded model because the fields being guarded are
// `TypeFloat`: after decoding, a `threshold` of `0` is indistinguishable from an absent one, and `0`
// is a legitimate value for the static operators. Only the raw config preserves the null.
//
// The descent is block-aware rather than name-aware. A nested block always appears in the raw
// configuration as a list or set of objects hanging off an object, so only values of that shape are
// descended into or dispatched to a validator. Matching on the name alone is not safe, because a
// `tags` entry is a map with user-controlled keys: a tag literally named `dependency` would
// otherwise hand a `cty.String` to a validator that immediately calls `LengthInt`, crashing the
// provider. Restricting the walk to block-shaped values removes that possibility rather than
// special-casing `tags`, and costs nothing, because a map can only ever hold primitives and so can
// never contain a block.
func walkRawConfigBlocks(input cty.Value, validators map[string]func(cty.Value) error) error {
	if input.IsNull() || !input.IsKnown() {
		return nil
	}

	switch t := input.Type(); {
	case t.IsObjectType():
		if input.LengthInt() == 0 {
			return nil
		}

		for key, value := range input.AsValueMap() {
			if !isRawConfigBlock(value) {
				continue
			}

			if validator, ok := validators[key]; ok {
				if err := validator(value); err != nil {
					return err
				}
			}

			if err := walkRawConfigBlocks(value, validators); err != nil {
				return err
			}
		}

	case t.IsListType(), t.IsSetType(), t.IsTupleType():
		if input.LengthInt() == 0 {
			return nil
		}

		for _, value := range input.AsValueSlice() {
			if err := walkRawConfigBlocks(value, validators); err != nil {
				return err
			}
		}
	}

	return nil
}

// isRawConfigBlock reports whether a raw configuration value has the shape the plugin SDK gives a
// nested block: a list or set of objects. Terraform models a `TypeList`/`TypeSet` block that way
// regardless of `MaxItems`, so a `MaxItems: 1` block is still a single-element list here.
func isRawConfigBlock(input cty.Value) bool {
	if input.IsNull() || !input.IsKnown() {
		return false
	}

	switch t := input.Type(); {
	case t.IsListType(), t.IsSetType():
		return t.ElementType().IsObjectType()

	case t.IsTupleType():
		for _, elementType := range t.TupleElementTypes() {
			if !elementType.IsObjectType() {
				return false
			}
		}

		return true
	}

	return false
}

// validateEvaluationRuleRawConfig rejects the `evaluation_rule` combinations the service silently
// ignores. `threshold` is only honoured for the static operators and `sensitivity`
// only for `Dynamic`, so a configuration that sets the wrong one never
// converges: the value is dropped on the way out and comes back absent on the next read, leaving a
// permanent diff.
func validateEvaluationRuleRawConfig(input cty.Value) error {
	return walkRawConfigBlocks(input, map[string]func(cty.Value) error{
		"evaluation_rule": validateEvaluationRuleBlock,
	})
}

func validateEvaluationRuleBlock(input cty.Value) error {
	if input.IsNull() || !input.IsKnown() || input.LengthInt() == 0 {
		return nil
	}

	for _, rule := range input.AsValueSlice() {
		if rule.IsNull() || !rule.IsKnown() || !rule.Type().IsObjectType() {
			continue
		}

		for _, blockName := range []string{"unhealthy", "degraded"} {
			if !rule.Type().HasAttribute(blockName) {
				continue
			}

			block := rule.GetAttr(blockName)
			if block.IsNull() || !block.IsKnown() || block.LengthInt() == 0 {
				continue
			}

			for _, thresholdRule := range block.AsValueSlice() {
				if err := validateThresholdRuleRawConfig(thresholdRule, blockName); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func validateThresholdRuleRawConfig(input cty.Value, blockName string) error {
	if input.IsNull() || !input.IsKnown() || !input.Type().IsObjectType() {
		return nil
	}

	operator := input.GetAttr("operator")
	if operator.IsNull() || !operator.IsKnown() {
		return nil
	}

	// An unknown value is still a value the user asked to send, so only an explicit null counts as
	// "not configured" here.
	isConfigured := func(name string) bool {
		if !input.Type().HasAttribute(name) {
			return false
		}
		return !input.GetAttr(name).IsNull()
	}

	if operator.AsString() == string(entities.SignalOperatorDynamic) {
		if isConfigured("threshold") {
			return fmt.Errorf("`threshold` cannot be set in `evaluation_rule.%s` when `operator` is `%s` - the service derives the threshold itself and only `sensitivity` applies", blockName, entities.SignalOperatorDynamic)
		}

		return nil
	}

	if isConfigured("sensitivity") {
		return fmt.Errorf("`sensitivity` cannot be set in `evaluation_rule.%s` when `operator` is `%s` - it only applies when `operator` is `%s`", blockName, operator.AsString(), entities.SignalOperatorDynamic)
	}

	if !isConfigured("threshold") {
		return fmt.Errorf("`threshold` must be set in `evaluation_rule.%s` when `operator` is `%s`", blockName, operator.AsString())
	}

	return nil
}

func validateEntityRawConfig(input cty.Value) error {
	validators := map[string]func(cty.Value) error{
		"signal_group": func(block cty.Value) error {
			return validatePopulatedBlock(block, "signal_group", []string{"azure_resource", "azure_log_analytics", "azure_monitor_workspace", "dependency"})
		},
		"alert": func(block cty.Value) error {
			return validatePopulatedBlock(block, "alert", []string{"unhealthy", "degraded"})
		},
	}
	for _, kind := range []string{"azure_resource", "azure_log_analytics", "azure_monitor_workspace"} {
		validators[kind] = func(block cty.Value) error {
			fields := []string{"query_text", "evaluation_rule"}
			if kind == "azure_resource" {
				fields = []string{"metric_namespace", "metric_name", "aggregation_type", "time_grain", "evaluation_rule"}
			}
			for _, group := range block.AsValueSlice() {
				if group.IsNull() || !group.IsKnown() || !group.Type().HasAttribute("signal") {
					continue
				}
				signals := group.GetAttr("signal")
				if signals.IsNull() || !signals.IsKnown() {
					continue
				}
				for _, signal := range signals.AsValueSlice() {
					if signal.IsNull() || !signal.IsKnown() {
						continue
					}
					if signal.Type().HasAttribute("signal_definition_name") && !signal.GetAttr("signal_definition_name").IsNull() {
						continue
					}
					for _, field := range fields {
						if !signal.Type().HasAttribute(field) || signal.GetAttr(field).IsNull() {
							return fmt.Errorf("`signal_group.%s.signal.%s` must be set when `signal_definition_name` is absent", kind, field)
						}
						value := signal.GetAttr(field)
						if value.IsKnown() && value.CanIterateElements() && value.LengthInt() == 0 {
							return fmt.Errorf("`signal_group.%s.signal.%s` must be set when `signal_definition_name` is absent", kind, field)
						}
					}
				}
			}
			return nil
		}
	}
	return walkRawConfigBlocks(input, validators)
}

func validatePopulatedBlock(input cty.Value, name string, children []string) error {
	for _, block := range input.AsValueSlice() {
		if block.IsNull() || !block.IsKnown() {
			continue
		}
		populated := false
		for _, child := range children {
			if !block.Type().HasAttribute(child) {
				continue
			}
			value := block.GetAttr(child)
			if !value.IsNull() && (!value.IsKnown() || value.LengthInt() > 0) {
				populated = true
			}
		}
		if !populated {
			return fmt.Errorf("`%s` must contain at least one of %v", name, children)
		}
	}
	return nil
}
