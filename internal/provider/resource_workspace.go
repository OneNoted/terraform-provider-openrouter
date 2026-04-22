package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/openrouter/terraform-provider-openrouter/internal/openrouter"
)

var (
	_ resource.Resource                = (*workspaceResource)(nil)
	_ resource.ResourceWithConfigure   = (*workspaceResource)(nil)
	_ resource.ResourceWithImportState = (*workspaceResource)(nil)
)

type workspaceResource struct {
	client *openrouter.Client
}

type workspaceModel struct {
	ID                              types.String `tfsdk:"id"`
	Name                            types.String `tfsdk:"name"`
	Slug                            types.String `tfsdk:"slug"`
	Description                     types.String `tfsdk:"description"`
	DefaultTextModel                types.String `tfsdk:"default_text_model"`
	DefaultImageModel               types.String `tfsdk:"default_image_model"`
	DefaultProviderSort             types.String `tfsdk:"default_provider_sort"`
	CreatedBy                       types.String `tfsdk:"created_by"`
	CreatedAt                       types.String `tfsdk:"created_at"`
	UpdatedAt                       types.String `tfsdk:"updated_at"`
	IsDataDiscountLoggingEnabled    types.Bool   `tfsdk:"is_data_discount_logging_enabled"`
	IsObservabilityBroadcastEnabled types.Bool   `tfsdk:"is_observability_broadcast_enabled"`
	IsObservabilityIOLoggingEnabled types.Bool   `tfsdk:"is_observability_io_logging_enabled"`
}

func NewWorkspaceResource() resource.Resource { return &workspaceResource{} }

func (r *workspaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace"
}

func (r *workspaceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *workspaceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		MarkdownDescription: "Manages an OpenRouter workspace.",
		Attributes: map[string]rschema.Attribute{
			"id":                                  rschema.StringAttribute{Computed: true, MarkdownDescription: "OpenRouter workspace UUID."},
			"name":                                rschema.StringAttribute{Required: true, MarkdownDescription: "Workspace display name."},
			"slug":                                rschema.StringAttribute{Required: true, MarkdownDescription: "URL-friendly workspace slug."},
			"description":                         rschema.StringAttribute{Optional: true, MarkdownDescription: "Workspace description."},
			"default_text_model":                  rschema.StringAttribute{Optional: true, MarkdownDescription: "Default text model for this workspace."},
			"default_image_model":                 rschema.StringAttribute{Optional: true, MarkdownDescription: "Default image model for this workspace."},
			"default_provider_sort":               rschema.StringAttribute{Optional: true, MarkdownDescription: "Default provider sort preference (`price`, `throughput`, `latency`, or `exacto`)."},
			"created_by":                          rschema.StringAttribute{Computed: true, MarkdownDescription: "OpenRouter user ID that created the workspace."},
			"created_at":                          rschema.StringAttribute{Computed: true, MarkdownDescription: "Creation timestamp returned by OpenRouter."},
			"updated_at":                          rschema.StringAttribute{Computed: true, MarkdownDescription: "Last update timestamp returned by OpenRouter."},
			"is_data_discount_logging_enabled":    rschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether data discount logging is enabled."},
			"is_observability_broadcast_enabled":  rschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether observability broadcast is enabled."},
			"is_observability_io_logging_enabled": rschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether private observability IO logging is enabled."},
		},
	}
}

func (r *workspaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		addProviderNotConfiguredError(&resp.Diagnostics)
		return
	}
	var plan workspaceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := workspaceRequestBody(plan)
	workspace, err := r.client.CreateWorkspace(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Create OpenRouter workspace failed", err.Error())
		return
	}
	state := workspaceModelFromAPI(ctx, *workspace, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *workspaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		addProviderNotConfiguredError(&resp.Diagnostics)
		return
	}
	var state workspaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := state.ID.ValueString()
	if id == "" {
		id = state.Slug.ValueString()
	}
	workspace, err := r.client.GetWorkspace(ctx, id)
	if openrouter.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read OpenRouter workspace failed", err.Error())
		return
	}
	newState := workspaceModelFromAPI(ctx, *workspace, state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *workspaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		addProviderNotConfiguredError(&resp.Diagnostics)
		return
	}
	var plan workspaceModel
	var state workspaceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	workspace, err := r.client.UpdateWorkspace(ctx, state.ID.ValueString(), workspaceUpdateRequestBody(plan, state))
	if err != nil {
		resp.Diagnostics.AddError("Update OpenRouter workspace failed", err.Error())
		return
	}
	newState := workspaceModelFromAPI(ctx, *workspace, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *workspaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		addProviderNotConfiguredError(&resp.Diagnostics)
		return
	}
	var state workspaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteWorkspace(ctx, state.ID.ValueString()); err != nil && !openrouter.IsNotFound(err) {
		resp.Diagnostics.AddError("Delete OpenRouter workspace failed", err.Error())
	}
}

func (r *workspaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func workspaceRequestBody(plan workspaceModel) map[string]any {
	body := map[string]any{}
	addStringIfKnown(body, "name", plan.Name)
	addStringIfKnown(body, "slug", plan.Slug)
	addStringIfKnown(body, "description", plan.Description)
	addStringIfKnown(body, "default_text_model", plan.DefaultTextModel)
	addStringIfKnown(body, "default_image_model", plan.DefaultImageModel)
	addStringIfKnown(body, "default_provider_sort", plan.DefaultProviderSort)
	return body
}

func workspaceUpdateRequestBody(plan, prior workspaceModel) map[string]any {
	body := map[string]any{}
	addStringIfKnown(body, "name", plan.Name)
	addStringIfKnown(body, "slug", plan.Slug)
	addNullableStringForUpdate(body, "description", plan.Description, prior.Description)
	addNullableStringForUpdate(body, "default_text_model", plan.DefaultTextModel, prior.DefaultTextModel)
	addNullableStringForUpdate(body, "default_image_model", plan.DefaultImageModel, prior.DefaultImageModel)
	addNullableStringForUpdate(body, "default_provider_sort", plan.DefaultProviderSort, prior.DefaultProviderSort)
	return body
}

func workspaceModelFromAPI(_ context.Context, workspace openrouter.Workspace, prior workspaceModel) workspaceModel {
	model := workspaceModel{
		ID:                              types.StringValue(workspace.ID),
		Name:                            types.StringValue(workspace.Name),
		Slug:                            types.StringValue(workspace.Slug),
		Description:                     stringValue(workspace.Description),
		DefaultTextModel:                stringValue(workspace.DefaultTextModel),
		DefaultImageModel:               stringValue(workspace.DefaultImageModel),
		DefaultProviderSort:             stringValue(workspace.DefaultProviderSort),
		CreatedBy:                       stringValue(workspace.CreatedBy),
		CreatedAt:                       types.StringValue(workspace.CreatedAt),
		UpdatedAt:                       stringValue(workspace.UpdatedAt),
		IsDataDiscountLoggingEnabled:    boolValue(workspace.IsDataDiscountLoggingEnabled),
		IsObservabilityBroadcastEnabled: boolValue(workspace.IsObservabilityBroadcastEnabled),
		IsObservabilityIOLoggingEnabled: boolValue(workspace.IsObservabilityIOLoggingEnabled),
	}
	if model.ID.ValueString() == "" && !prior.ID.IsNull() {
		model.ID = prior.ID
	}
	return model
}
