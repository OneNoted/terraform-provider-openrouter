package provider

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func stringPointerValue(value types.String) *string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	v := value.ValueString()
	return &v
}

func floatPointerValue(value types.Float64) *float64 {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	v := value.ValueFloat64()
	return &v
}

func boolPointerValue(value types.Bool) *bool {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	v := value.ValueBool()
	return &v
}

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
	sorted := append([]string(nil), values...)
	sort.Strings(sorted)
	return types.ListValueFrom(ctx, types.StringType, sorted)
}

func setStringsFromList(ctx context.Context, list types.List) ([]string, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var values []string
	diags := list.ElementsAs(ctx, &values, false)
	sort.Strings(values)
	return values, diags
}

func addStringIfKnown(body map[string]any, key string, value types.String) {
	if value.IsUnknown() {
		return
	}
	if value.IsNull() {
		body[key] = nil
		return
	}
	body[key] = value.ValueString()
}

func addFloatIfKnown(body map[string]any, key string, value types.Float64) {
	if value.IsUnknown() {
		return
	}
	if value.IsNull() {
		body[key] = nil
		return
	}
	body[key] = value.ValueFloat64()
}

func addBoolIfKnown(body map[string]any, key string, value types.Bool) {
	if value.IsUnknown() {
		return
	}
	if value.IsNull() {
		body[key] = nil
		return
	}
	body[key] = value.ValueBool()
}

func addListIfKnown(ctx context.Context, body map[string]any, key string, value types.List) diag.Diagnostics {
	if value.IsUnknown() {
		return nil
	}
	if value.IsNull() {
		body[key] = nil
		return nil
	}
	values, diags := setStringsFromList(ctx, value)
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
