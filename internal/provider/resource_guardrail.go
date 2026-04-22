package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/openrouter/terraform-provider-openrouter/internal/openrouter"
)

var (
	_ resource.Resource                = (*guardrailResource)(nil)
	_ resource.ResourceWithConfigure   = (*guardrailResource)(nil)
	_ resource.ResourceWithImportState = (*guardrailResource)(nil)
)

type guardrailResource struct {
	client *openrouter.Client
}

type guardrailModel struct {
	ID               types.String  `tfsdk:"id"`
	Name             types.String  `tfsdk:"name"`
	WorkspaceID      types.String  `tfsdk:"workspace_id"`
	Description      types.String  `tfsdk:"description"`
	AllowedModels    types.Set     `tfsdk:"allowed_models"`
	AllowedProviders types.Set     `tfsdk:"allowed_providers"`
	IgnoredModels    types.Set     `tfsdk:"ignored_models"`
	IgnoredProviders types.Set     `tfsdk:"ignored_providers"`
	EnforceZDR       types.Bool    `tfsdk:"enforce_zdr"`
	LimitUSD         types.Float64 `tfsdk:"limit_usd"`
	ResetInterval    types.String  `tfsdk:"reset_interval"`
	CreatedAt        types.String  `tfsdk:"created_at"`
	UpdatedAt        types.String  `tfsdk:"updated_at"`
}

func NewGuardrailResource() resource.Resource { return &guardrailResource{} }

func (r *guardrailResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_guardrail"
}

func (r *guardrailResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*openrouter.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", "Expected *openrouter.Client.")
		return
	}
	r.client = client
}

func (r *guardrailResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		MarkdownDescription: "Manages an OpenRouter guardrail for provider/model restrictions, ZDR enforcement, and budget limits.",
		Attributes: map[string]rschema.Attribute{
			"id":                rschema.StringAttribute{Computed: true, MarkdownDescription: "OpenRouter guardrail UUID."},
			"name":              rschema.StringAttribute{Required: true, MarkdownDescription: "Guardrail name."},
			"workspace_id":      rschema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Workspace ID containing the guardrail. Defaults to OpenRouter's default workspace if omitted."},
			"description":       rschema.StringAttribute{Optional: true, MarkdownDescription: "Guardrail description."},
			"allowed_models":    rschema.SetAttribute{Optional: true, ElementType: types.StringType, MarkdownDescription: "Optional allowlist of model slugs/canonical slugs."},
			"allowed_providers": rschema.SetAttribute{Optional: true, ElementType: types.StringType, MarkdownDescription: "Optional allowlist of provider IDs."},
			"ignored_models":    rschema.SetAttribute{Optional: true, ElementType: types.StringType, MarkdownDescription: "Optional list of model slugs/canonical slugs to exclude from routing."},
			"ignored_providers": rschema.SetAttribute{Optional: true, ElementType: types.StringType, MarkdownDescription: "Optional list of provider IDs to exclude from routing."},
			"enforce_zdr":       rschema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Whether zero-data-retention providers are required."},
			"limit_usd":         rschema.Float64Attribute{Optional: true, MarkdownDescription: "Spending limit in USD."},
			"reset_interval":    rschema.StringAttribute{Optional: true, MarkdownDescription: "Budget reset interval (`daily`, `weekly`, or `monthly`)."},
			"created_at":        rschema.StringAttribute{Computed: true, MarkdownDescription: "Creation timestamp returned by OpenRouter."},
			"updated_at":        rschema.StringAttribute{Computed: true, MarkdownDescription: "Last update timestamp returned by OpenRouter."},
		},
	}
}

func (r *guardrailResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan guardrailModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body, diags := guardrailRequestBody(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	guardrail, err := r.client.CreateGuardrail(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Create OpenRouter guardrail failed", err.Error())
		return
	}
	state := guardrailModelFromAPI(ctx, *guardrail, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *guardrailResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state guardrailModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	guardrail, err := r.client.GetGuardrail(ctx, state.ID.ValueString())
	if openrouter.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read OpenRouter guardrail failed", err.Error())
		return
	}
	newState := guardrailModelFromAPI(ctx, *guardrail, state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *guardrailResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan guardrailModel
	var state guardrailModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body, diags := guardrailUpdateRequestBody(ctx, plan, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	guardrail, err := r.client.UpdateGuardrail(ctx, state.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Update OpenRouter guardrail failed", err.Error())
		return
	}
	newState := guardrailModelFromAPI(ctx, *guardrail, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *guardrailResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state guardrailModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteGuardrail(ctx, state.ID.ValueString()); err != nil && !openrouter.IsNotFound(err) {
		resp.Diagnostics.AddError("Delete OpenRouter guardrail failed", err.Error())
	}
}

func (r *guardrailResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func guardrailRequestBody(ctx context.Context, plan guardrailModel) (map[string]any, diag.Diagnostics) {
	body := map[string]any{}
	addStringIfKnown(body, "name", plan.Name)
	addStringIfKnown(body, "workspace_id", plan.WorkspaceID)
	addStringIfKnown(body, "description", plan.Description)
	addBoolIfKnown(body, "enforce_zdr", plan.EnforceZDR)
	addFloatIfKnown(body, "limit_usd", plan.LimitUSD)
	addStringIfKnown(body, "reset_interval", plan.ResetInterval)
	var diags diag.Diagnostics
	diags.Append(addSetIfKnown(ctx, body, "allowed_models", plan.AllowedModels)...)
	diags.Append(addSetIfKnown(ctx, body, "allowed_providers", plan.AllowedProviders)...)
	diags.Append(addSetIfKnown(ctx, body, "ignored_models", plan.IgnoredModels)...)
	diags.Append(addSetIfKnown(ctx, body, "ignored_providers", plan.IgnoredProviders)...)
	return body, diags
}

func guardrailUpdateRequestBody(ctx context.Context, plan, prior guardrailModel) (map[string]any, diag.Diagnostics) {
	body := map[string]any{}
	addStringIfKnown(body, "name", plan.Name)
	addStringIfKnown(body, "workspace_id", plan.WorkspaceID)
	addNullableStringForUpdate(body, "description", plan.Description, prior.Description)
	addNullableBoolForUpdate(body, "enforce_zdr", plan.EnforceZDR, prior.EnforceZDR)
	addNullableFloatForUpdate(body, "limit_usd", plan.LimitUSD, prior.LimitUSD)
	addStringIfKnown(body, "reset_interval", plan.ResetInterval)
	var diags diag.Diagnostics
	diags.Append(addNullableSetForUpdate(ctx, body, "allowed_models", plan.AllowedModels, prior.AllowedModels)...)
	diags.Append(addNullableSetForUpdate(ctx, body, "allowed_providers", plan.AllowedProviders, prior.AllowedProviders)...)
	diags.Append(addNullableSetForUpdate(ctx, body, "ignored_models", plan.IgnoredModels, prior.IgnoredModels)...)
	diags.Append(addNullableSetForUpdate(ctx, body, "ignored_providers", plan.IgnoredProviders, prior.IgnoredProviders)...)
	return body, diags
}

func guardrailModelFromAPI(ctx context.Context, guardrail openrouter.Guardrail, prior guardrailModel) guardrailModel {
	allowedModels, diags := stringSetValue(ctx, guardrail.AllowedModels)
	if diags.HasError() {
		allowedModels = prior.AllowedModels
	}
	allowedProviders, diags := stringSetValue(ctx, guardrail.AllowedProviders)
	if diags.HasError() {
		allowedProviders = prior.AllowedProviders
	}
	ignoredModels, diags := stringSetValue(ctx, guardrail.IgnoredModels)
	if diags.HasError() {
		ignoredModels = prior.IgnoredModels
	}
	ignoredProviders, diags := stringSetValue(ctx, guardrail.IgnoredProviders)
	if diags.HasError() {
		ignoredProviders = prior.IgnoredProviders
	}
	workspaceID := stringValue(guardrail.WorkspaceID)
	if workspaceID.IsNull() && !prior.WorkspaceID.IsUnknown() {
		workspaceID = prior.WorkspaceID
	}
	return guardrailModel{
		ID:               types.StringValue(guardrail.ID),
		Name:             types.StringValue(guardrail.Name),
		WorkspaceID:      workspaceID,
		Description:      stringValue(guardrail.Description),
		AllowedModels:    allowedModels,
		AllowedProviders: allowedProviders,
		IgnoredModels:    ignoredModels,
		IgnoredProviders: ignoredProviders,
		EnforceZDR:       boolValue(guardrail.EnforceZDR),
		LimitUSD:         floatValue(guardrail.LimitUSD),
		ResetInterval:    stringValue(guardrail.ResetInterval),
		CreatedAt:        types.StringValue(guardrail.CreatedAt),
		UpdatedAt:        types.StringValue(guardrail.UpdatedAt),
	}
}
