package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestOptionalComputedResourceAttributesUsePlanModifiers(t *testing.T) {
	ctx := context.Background()

	apiKey := NewAPIKeyResource()
	var apiKeySchema resource.SchemaResponse
	apiKey.Schema(ctx, resource.SchemaRequest{}, &apiKeySchema)
	if apiKeySchema.Diagnostics.HasError() {
		t.Fatalf("api key schema diagnostics: %v", apiKeySchema.Diagnostics)
	}
	assertBoolPlanModifier(t, apiKeySchema.Schema.Attributes, "disabled")
	assertBoolPlanModifier(t, apiKeySchema.Schema.Attributes, "include_byok_in_limit")
	assertStringPlanModifier(t, apiKeySchema.Schema.Attributes, "workspace_id")
	assertStringPlanModifier(t, apiKeySchema.Schema.Attributes, "expires_at")

	guardrail := NewGuardrailResource()
	var guardrailSchema resource.SchemaResponse
	guardrail.Schema(ctx, resource.SchemaRequest{}, &guardrailSchema)
	if guardrailSchema.Diagnostics.HasError() {
		t.Fatalf("guardrail schema diagnostics: %v", guardrailSchema.Diagnostics)
	}
	assertStringPlanModifier(t, guardrailSchema.Schema.Attributes, "workspace_id")
	assertBoolPlanModifier(t, guardrailSchema.Schema.Attributes, "enforce_zdr")
}

func TestResourcesReturnDiagnosticsWhenProviderClientMissing(t *testing.T) {
	ctx := context.Background()
	for name, res := range map[string]resource.Resource{
		"workspace": NewWorkspaceResource(),
		"api_key":   NewAPIKeyResource(),
		"guardrail": NewGuardrailResource(),
	} {
		t.Run(name+" create", func(t *testing.T) {
			var resp resource.CreateResponse
			res.Create(ctx, resource.CreateRequest{}, &resp)
			assertProviderNotConfigured(t, resp.Diagnostics)
		})
		t.Run(name+" read", func(t *testing.T) {
			var resp resource.ReadResponse
			res.Read(ctx, resource.ReadRequest{}, &resp)
			assertProviderNotConfigured(t, resp.Diagnostics)
		})
		t.Run(name+" update", func(t *testing.T) {
			var resp resource.UpdateResponse
			res.Update(ctx, resource.UpdateRequest{}, &resp)
			assertProviderNotConfigured(t, resp.Diagnostics)
		})
		t.Run(name+" delete", func(t *testing.T) {
			var resp resource.DeleteResponse
			res.Delete(ctx, resource.DeleteRequest{}, &resp)
			assertProviderNotConfigured(t, resp.Diagnostics)
		})
	}
}

func TestDataSourcesReturnDiagnosticsWhenProviderClientMissing(t *testing.T) {
	ctx := context.Background()
	for name, ds := range map[string]datasource.DataSource{
		"providers":  NewProvidersDataSource(),
		"models":     NewModelsDataSource(),
		"workspaces": NewWorkspacesDataSource(),
	} {
		t.Run(name, func(t *testing.T) {
			var resp datasource.ReadResponse
			ds.Read(ctx, datasource.ReadRequest{}, &resp)
			assertProviderNotConfigured(t, resp.Diagnostics)
		})
	}
}

func assertBoolPlanModifier(t *testing.T, attrs map[string]rschema.Attribute, name string) {
	t.Helper()
	attr, ok := attrs[name].(rschema.BoolAttribute)
	if !ok {
		t.Fatalf("%s is %T, want BoolAttribute", name, attrs[name])
	}
	if len(attr.PlanModifiers) == 0 {
		t.Fatalf("%s should have a plan modifier", name)
	}
}

func assertStringPlanModifier(t *testing.T, attrs map[string]rschema.Attribute, name string) {
	t.Helper()
	attr, ok := attrs[name].(rschema.StringAttribute)
	if !ok {
		t.Fatalf("%s is %T, want StringAttribute", name, attrs[name])
	}
	if len(attr.PlanModifiers) == 0 {
		t.Fatalf("%s should have a plan modifier", name)
	}
}

func assertProviderNotConfigured(t *testing.T, diagnostics interface{ HasError() bool }) {
	t.Helper()
	if !diagnostics.HasError() {
		t.Fatal("expected provider-not-configured diagnostic")
	}
}
