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
	_ resource.Resource                = (*apiKeyResource)(nil)
	_ resource.ResourceWithConfigure   = (*apiKeyResource)(nil)
	_ resource.ResourceWithImportState = (*apiKeyResource)(nil)
)

type apiKeyResource struct {
	client *openrouter.Client
}

type apiKeyModel struct {
	ID                 types.String  `tfsdk:"id"`
	Hash               types.String  `tfsdk:"hash"`
	Name               types.String  `tfsdk:"name"`
	Label              types.String  `tfsdk:"label"`
	WorkspaceID        types.String  `tfsdk:"workspace_id"`
	Disabled           types.Bool    `tfsdk:"disabled"`
	Limit              types.Float64 `tfsdk:"limit"`
	LimitRemaining     types.Float64 `tfsdk:"limit_remaining"`
	LimitReset         types.String  `tfsdk:"limit_reset"`
	IncludeBYOKInLimit types.Bool    `tfsdk:"include_byok_in_limit"`
	Usage              types.Float64 `tfsdk:"usage"`
	UsageDaily         types.Float64 `tfsdk:"usage_daily"`
	UsageWeekly        types.Float64 `tfsdk:"usage_weekly"`
	UsageMonthly       types.Float64 `tfsdk:"usage_monthly"`
	BYOKUsage          types.Float64 `tfsdk:"byok_usage"`
	BYOKUsageDaily     types.Float64 `tfsdk:"byok_usage_daily"`
	BYOKUsageWeekly    types.Float64 `tfsdk:"byok_usage_weekly"`
	BYOKUsageMonthly   types.Float64 `tfsdk:"byok_usage_monthly"`
	CreatedAt          types.String  `tfsdk:"created_at"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	ExpiresAt          types.String  `tfsdk:"expires_at"`
	CreatorUserID      types.String  `tfsdk:"creator_user_id"`
}

func NewAPIKeyResource() resource.Resource { return &apiKeyResource{} }

func (r *apiKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (r *apiKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *apiKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		MarkdownDescription: "Manages an OpenRouter API key. The one-time plaintext key returned by OpenRouter on create is deliberately discarded and never stored in Terraform state.",
		Attributes: map[string]rschema.Attribute{
			"id":                    rschema.StringAttribute{Computed: true, MarkdownDescription: "Stable API key identifier. Same value as `hash`."},
			"hash":                  rschema.StringAttribute{Computed: true, MarkdownDescription: "Stable OpenRouter API key hash/identifier."},
			"name":                  rschema.StringAttribute{Required: true, MarkdownDescription: "API key display name."},
			"label":                 rschema.StringAttribute{Computed: true, MarkdownDescription: "OpenRouter label for the key."},
			"workspace_id":          rschema.StringAttribute{Optional: true, MarkdownDescription: "Optional workspace ID for APIs/accounts that support workspace-scoped key creation."},
			"disabled":              rschema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Whether the key is disabled."},
			"limit":                 rschema.Float64Attribute{Optional: true, MarkdownDescription: "Optional spending limit in USD."},
			"limit_remaining":       rschema.Float64Attribute{Computed: true, MarkdownDescription: "Remaining limit returned by OpenRouter."},
			"limit_reset":           rschema.StringAttribute{Optional: true, MarkdownDescription: "Limit reset interval (`daily`, `weekly`, `monthly`) or null for no reset."},
			"include_byok_in_limit": rschema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Whether BYOK usage counts against the key limit."},
			"usage":                 rschema.Float64Attribute{Computed: true, MarkdownDescription: "Total usage returned by OpenRouter."},
			"usage_daily":           rschema.Float64Attribute{Computed: true, MarkdownDescription: "Daily usage returned by OpenRouter."},
			"usage_weekly":          rschema.Float64Attribute{Computed: true, MarkdownDescription: "Weekly usage returned by OpenRouter."},
			"usage_monthly":         rschema.Float64Attribute{Computed: true, MarkdownDescription: "Monthly usage returned by OpenRouter."},
			"byok_usage":            rschema.Float64Attribute{Computed: true, MarkdownDescription: "Total BYOK usage returned by OpenRouter."},
			"byok_usage_daily":      rschema.Float64Attribute{Computed: true, MarkdownDescription: "Daily BYOK usage returned by OpenRouter."},
			"byok_usage_weekly":     rschema.Float64Attribute{Computed: true, MarkdownDescription: "Weekly BYOK usage returned by OpenRouter."},
			"byok_usage_monthly":    rschema.Float64Attribute{Computed: true, MarkdownDescription: "Monthly BYOK usage returned by OpenRouter."},
			"created_at":            rschema.StringAttribute{Computed: true, MarkdownDescription: "Creation timestamp returned by OpenRouter."},
			"updated_at":            rschema.StringAttribute{Computed: true, MarkdownDescription: "Last update timestamp returned by OpenRouter."},
			"expires_at":            rschema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional UTC expiration timestamp."},
			"creator_user_id":       rschema.StringAttribute{Computed: true, MarkdownDescription: "Creator user ID returned by OpenRouter."},
		},
	}
}

func (r *apiKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan apiKeyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiKey, err := r.client.CreateAPIKey(ctx, apiKeyRequestBody(plan))
	if err != nil {
		resp.Diagnostics.AddError("Create OpenRouter API key failed", err.Error())
		return
	}
	state := apiKeyModelFromAPI(*apiKey, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *apiKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state apiKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiKey, err := r.client.GetAPIKey(ctx, keyHash(state))
	if openrouter.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read OpenRouter API key failed", err.Error())
		return
	}
	newState := apiKeyModelFromAPI(*apiKey, state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *apiKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan apiKeyModel
	var state apiKeyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiKey, err := r.client.UpdateAPIKey(ctx, keyHash(state), apiKeyRequestBody(plan))
	if err != nil {
		resp.Diagnostics.AddError("Update OpenRouter API key failed", err.Error())
		return
	}
	newState := apiKeyModelFromAPI(*apiKey, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *apiKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state apiKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteAPIKey(ctx, keyHash(state)); err != nil && !openrouter.IsNotFound(err) {
		resp.Diagnostics.AddError("Delete OpenRouter API key failed", err.Error())
	}
}

func (r *apiKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func apiKeyRequestBody(plan apiKeyModel) map[string]any {
	body := map[string]any{}
	addStringIfKnown(body, "name", plan.Name)
	addStringIfKnown(body, "workspace_id", plan.WorkspaceID)
	addBoolIfKnown(body, "disabled", plan.Disabled)
	addFloatIfKnown(body, "limit", plan.Limit)
	addStringIfKnown(body, "limit_reset", plan.LimitReset)
	addBoolIfKnown(body, "include_byok_in_limit", plan.IncludeBYOKInLimit)
	addStringIfKnown(body, "expires_at", plan.ExpiresAt)
	return body
}

func keyHash(model apiKeyModel) string {
	if !model.Hash.IsNull() && model.Hash.ValueString() != "" {
		return model.Hash.ValueString()
	}
	return model.ID.ValueString()
}

func apiKeyModelFromAPI(apiKey openrouter.APIKey, prior apiKeyModel) apiKeyModel {
	workspaceID := stringValue(apiKey.WorkspaceID)
	if workspaceID.IsNull() && !prior.WorkspaceID.IsUnknown() {
		workspaceID = prior.WorkspaceID
	}
	expiresAt := stringValue(apiKey.ExpiresAt)
	if expiresAt.IsNull() && !prior.ExpiresAt.IsUnknown() {
		expiresAt = prior.ExpiresAt
	}
	return apiKeyModel{
		ID:                 types.StringValue(apiKey.Hash),
		Hash:               types.StringValue(apiKey.Hash),
		Name:               types.StringValue(apiKey.Name),
		Label:              types.StringValue(apiKey.Label),
		WorkspaceID:        workspaceID,
		Disabled:           boolValue(apiKey.Disabled),
		Limit:              floatValue(apiKey.Limit),
		LimitRemaining:     floatValue(apiKey.LimitRemaining),
		LimitReset:         stringValue(apiKey.LimitReset),
		IncludeBYOKInLimit: boolValue(apiKey.IncludeBYOKInLimit),
		Usage:              floatValue(apiKey.Usage),
		UsageDaily:         floatValue(apiKey.UsageDaily),
		UsageWeekly:        floatValue(apiKey.UsageWeekly),
		UsageMonthly:       floatValue(apiKey.UsageMonthly),
		BYOKUsage:          floatValue(apiKey.BYOKUsage),
		BYOKUsageDaily:     floatValue(apiKey.BYOKUsageDaily),
		BYOKUsageWeekly:    floatValue(apiKey.BYOKUsageWeekly),
		BYOKUsageMonthly:   floatValue(apiKey.BYOKUsageMonthly),
		CreatedAt:          types.StringValue(apiKey.CreatedAt),
		UpdatedAt:          types.StringValue(apiKey.UpdatedAt),
		ExpiresAt:          expiresAt,
		CreatorUserID:      types.StringValue(apiKey.CreatorUserID),
	}
}
