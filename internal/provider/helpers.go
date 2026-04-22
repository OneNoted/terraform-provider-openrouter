package provider

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func stringValue(value *string) types.String {
	if value == nil {
		return types.StringNull()
	}
	return types.StringValue(*value)
}

func floatValue(value *float64) types.Float64 {
	if value == nil {
		return types.Float64Null()
	}
	return types.Float64Value(*value)
}

func boolValue(value *bool) types.Bool {
	if value == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*value)
}

func stringListValue(ctx context.Context, values []string) (types.List, diag.Diagnostics) {
	if values == nil {
		return types.ListNull(types.StringType), nil
	}
	return types.ListValueFrom(ctx, types.StringType, values)
}

func stringSetValue(ctx context.Context, values []string) (types.Set, diag.Diagnostics) {
	if values == nil {
		return types.SetNull(types.StringType), nil
	}
	return types.SetValueFrom(ctx, types.StringType, values)
}

func stringsFromSet(ctx context.Context, set types.Set) ([]string, diag.Diagnostics) {
	if set.IsNull() || set.IsUnknown() {
		return nil, nil
	}
	var values []string
	diags := set.ElementsAs(ctx, &values, false)
	sort.Strings(values)
	return values, diags
}

func addStringIfKnown(body map[string]any, key string, value types.String) {
	if value.IsUnknown() {
		return
	}
	if value.IsNull() {
		return
	}
	body[key] = value.ValueString()
}

func addFloatIfKnown(body map[string]any, key string, value types.Float64) {
	if value.IsUnknown() {
		return
	}
	if value.IsNull() {
		return
	}
	body[key] = value.ValueFloat64()
}

func addBoolIfKnown(body map[string]any, key string, value types.Bool) {
	if value.IsUnknown() {
		return
	}
	if value.IsNull() {
		return
	}
	body[key] = value.ValueBool()
}

func addSetIfKnown(ctx context.Context, body map[string]any, key string, value types.Set) diag.Diagnostics {
	if value.IsUnknown() || value.IsNull() {
		return nil
	}
	values, diags := stringsFromSet(ctx, value)
	if diags.HasError() {
		return diags
	}
	body[key] = values
	return nil
}

func jsonString(value map[string]any) types.String {
	if value == nil {
		return types.StringNull()
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return types.StringNull()
	}
	return types.StringValue(string(payload))
}

func addNullableStringForUpdate(body map[string]any, key string, plan, prior types.String) {
	if plan.IsUnknown() {
		return
	}
	if plan.IsNull() {
		if !prior.IsNull() && !prior.IsUnknown() {
			body[key] = nil
		}
		return
	}
	body[key] = plan.ValueString()
}

func addNullableFloatForUpdate(body map[string]any, key string, plan, prior types.Float64) {
	if plan.IsUnknown() {
		return
	}
	if plan.IsNull() {
		if !prior.IsNull() && !prior.IsUnknown() {
			body[key] = nil
		}
		return
	}
	body[key] = plan.ValueFloat64()
}

func addNullableBoolForUpdate(body map[string]any, key string, plan, prior types.Bool) {
	if plan.IsUnknown() {
		return
	}
	if plan.IsNull() {
		if !prior.IsNull() && !prior.IsUnknown() {
			body[key] = nil
		}
		return
	}
	body[key] = plan.ValueBool()
}

func addNullableSetForUpdate(ctx context.Context, body map[string]any, key string, plan, prior types.Set) diag.Diagnostics {
	if plan.IsUnknown() {
		return nil
	}
	if plan.IsNull() {
		if !prior.IsNull() && !prior.IsUnknown() {
			body[key] = nil
		}
		return nil
	}
	values, diags := stringsFromSet(ctx, plan)
	if diags.HasError() {
		return diags
	}
	body[key] = values
	return nil
}

func addProviderNotConfiguredError(diags *diag.Diagnostics) {
	diags.AddError(
		"OpenRouter provider is not configured",
		"The provider client was not available for this operation. Configure the openrouter provider with management_api_key or OPENROUTER_MANAGEMENT_API_KEY before using resources or data sources.",
	)
}
