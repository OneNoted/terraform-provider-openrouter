package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/openrouter/terraform-provider-openrouter/internal/openrouter"
)

func TestNullableAPIKeyFieldsStayNullInTerraformState(t *testing.T) {
	model := apiKeyModelFromAPI(openrouter.APIKey{
		Hash: "hash",
		Name: "name",
	}, apiKeyModel{})
	if !model.UpdatedAt.IsNull() {
		t.Fatalf("updated_at = %s, want null", model.UpdatedAt.String())
	}
	if !model.CreatorUserID.IsNull() {
		t.Fatalf("creator_user_id = %s, want null", model.CreatorUserID.String())
	}
}

func TestNullableWorkspaceFieldsStayNullInTerraformState(t *testing.T) {
	model := workspaceModelFromAPI(context.Background(), openrouter.Workspace{
		ID:   "workspace-id",
		Name: "workspace",
		Slug: "workspace",
	}, workspaceModel{})
	if !model.CreatedBy.IsNull() {
		t.Fatalf("created_by = %s, want null", model.CreatedBy.String())
	}
	if !model.UpdatedAt.IsNull() {
		t.Fatalf("updated_at = %s, want null", model.UpdatedAt.String())
	}
}

func TestNullableGuardrailFieldsStayNullInTerraformState(t *testing.T) {
	model := guardrailModelFromAPI(context.Background(), openrouter.Guardrail{
		ID:   "guardrail-id",
		Name: "guardrail",
	}, guardrailModel{})
	if !model.UpdatedAt.IsNull() {
		t.Fatalf("updated_at = %s, want null", model.UpdatedAt.String())
	}
}

func TestNullableProviderMetadataUsesTerraformNulls(t *testing.T) {
	if got := stringValue(nil); !got.Equal(types.StringNull()) {
		t.Fatalf("stringValue(nil) = %s, want null", got.String())
	}
}

func TestNullableModelFieldsStayNullInTerraformState(t *testing.T) {
	model := openrouter.Model{
		ID:            "model/id",
		Name:          "Model",
		CanonicalSlug: "model/id",
	}
	inputModalities, diags := stringListValue(context.Background(), model.Architecture.InputModalities)
	if diags.HasError() {
		t.Fatalf("diagnostics: %v", diags)
	}
	outputModalities, diags := stringListValue(context.Background(), model.Architecture.OutputModalities)
	if diags.HasError() {
		t.Fatalf("diagnostics: %v", diags)
	}
	mapped := modelItemModel{
		HuggingFaceID:    stringValue(model.HuggingFaceID),
		InstructType:     stringValue(model.Architecture.InstructType),
		InputModalities:  inputModalities,
		OutputModalities: outputModalities,
	}
	if !mapped.HuggingFaceID.IsNull() {
		t.Fatalf("hugging_face_id = %s, want null", mapped.HuggingFaceID.String())
	}
	if !mapped.InstructType.IsNull() {
		t.Fatalf("instruct_type = %s, want null", mapped.InstructType.String())
	}
}
