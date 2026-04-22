package provider

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestAPIKeyRequestBodyOmitsUnconfiguredOptionalFields(t *testing.T) {
	body := apiKeyRequestBody(apiKeyModel{Name: types.StringValue("ci")})
	want := map[string]any{"name": "ci"}
	if !reflect.DeepEqual(body, want) {
		t.Fatalf("body = %#v, want %#v", body, want)
	}
}

func TestAPIKeyCreateRequestBodyIncludesWorkspaceID(t *testing.T) {
	body := apiKeyRequestBody(apiKeyModel{
		Name:        types.StringValue("ci"),
		WorkspaceID: types.StringValue("workspace-id"),
	})
	if got, want := body["workspace_id"], "workspace-id"; got != want {
		t.Fatalf("workspace_id = %#v, want %#v", got, want)
	}
}

func TestGuardrailRequestBodyOmitsUnconfiguredOptionalFields(t *testing.T) {
	body, diags := guardrailRequestBody(context.Background(), guardrailModel{Name: types.StringValue("prod")})
	if diags.HasError() {
		t.Fatalf("diagnostics: %v", diags)
	}
	want := map[string]any{"name": "prod"}
	if !reflect.DeepEqual(body, want) {
		t.Fatalf("body = %#v, want %#v", body, want)
	}
}

func TestGuardrailRequestBodyUsesSetSemanticsForUnorderedCollections(t *testing.T) {
	providers, diags := types.SetValueFrom(context.Background(), types.StringType, []string{"openai", "anthropic"})
	if diags.HasError() {
		t.Fatalf("diagnostics: %v", diags)
	}
	body, diags := guardrailRequestBody(context.Background(), guardrailModel{
		Name:             types.StringValue("prod"),
		AllowedProviders: providers,
	})
	if diags.HasError() {
		t.Fatalf("diagnostics: %v", diags)
	}
	got, ok := body["allowed_providers"].([]string)
	if !ok {
		t.Fatalf("allowed_providers = %#v, want []string", body["allowed_providers"])
	}
	want := []string{"anthropic", "openai"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("allowed_providers = %#v, want sorted stable %#v", got, want)
	}
}

func TestWorkspaceUpdateRequestBodyClearsNullableFields(t *testing.T) {
	body := workspaceUpdateRequestBody(
		workspaceModel{Name: types.StringValue("prod"), Slug: types.StringValue("prod")},
		workspaceModel{Description: types.StringValue("old"), DefaultTextModel: types.StringValue("openai/gpt-4o")},
	)
	if got, ok := body["description"]; !ok || got != nil {
		t.Fatalf("description clear = %#v, present=%t; want null", got, ok)
	}
	if got, ok := body["default_text_model"]; !ok || got != nil {
		t.Fatalf("default_text_model clear = %#v, present=%t; want null", got, ok)
	}
}

func TestAPIKeyUpdateRequestBodyClearsNullableFields(t *testing.T) {
	body := apiKeyUpdateRequestBody(
		apiKeyModel{Name: types.StringValue("ci"), WorkspaceID: types.StringValue("workspace-id"), ExpiresAt: types.StringValue("2031-01-01T00:00:00Z")},
		apiKeyModel{WorkspaceID: types.StringValue("workspace-id"), Limit: types.Float64Value(10), LimitReset: types.StringValue("monthly"), ExpiresAt: types.StringValue("2030-01-01T00:00:00Z")},
	)
	for _, key := range []string{"limit", "limit_reset"} {
		if got, ok := body[key]; !ok || got != nil {
			t.Fatalf("%s clear = %#v, present=%t; want null", key, got, ok)
		}
	}
	if _, ok := body["workspace_id"]; ok {
		t.Fatalf("workspace_id should be omitted on update because OpenRouter does not support moving keys between workspaces")
	}
	if _, ok := body["expires_at"]; ok {
		t.Fatalf("expires_at should be omitted on update because OpenRouter does not support changing key expiration in-place")
	}
}

func TestGuardrailUpdateRequestBodyClearsNullableFields(t *testing.T) {
	oldProviders, diags := types.SetValueFrom(context.Background(), types.StringType, []string{"openai"})
	if diags.HasError() {
		t.Fatalf("diagnostics: %v", diags)
	}
	body, diags := guardrailUpdateRequestBody(context.Background(),
		guardrailModel{Name: types.StringValue("prod"), WorkspaceID: types.StringValue("workspace-id")},
		guardrailModel{WorkspaceID: types.StringValue("workspace-id"), Description: types.StringValue("old"), AllowedProviders: oldProviders, EnforceZDR: types.BoolValue(true), LimitUSD: types.Float64Value(100), ResetInterval: types.StringValue("monthly")},
	)
	if diags.HasError() {
		t.Fatalf("diagnostics: %v", diags)
	}
	for _, key := range []string{"description", "allowed_providers", "enforce_zdr", "limit_usd", "reset_interval"} {
		if got, ok := body[key]; !ok || got != nil {
			t.Fatalf("%s clear = %#v, present=%t; want null", key, got, ok)
		}
	}
	if _, ok := body["workspace_id"]; ok {
		t.Fatalf("workspace_id should be omitted on update because OpenRouter does not support moving guardrails between workspaces")
	}
}
