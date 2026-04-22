package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/openrouter/terraform-provider-openrouter/internal/openrouter"
)

var (
	_ datasource.DataSource              = (*workspacesDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*workspacesDataSource)(nil)
)

type workspacesDataSource struct{ client *openrouter.Client }

type workspacesDataSourceModel struct {
	Workspaces []workspaceDataModel `tfsdk:"workspaces"`
}

type workspaceDataModel struct {
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

func NewWorkspacesDataSource() datasource.DataSource { return &workspacesDataSource{} }

func (d *workspacesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspaces"
}

func (d *workspacesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*openrouter.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", "Expected *openrouter.Client.")
		return
	}
	d.client = client
}

func (d *workspacesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		MarkdownDescription: "Lists OpenRouter workspaces visible to the management key.",
		Attributes: map[string]dschema.Attribute{
			"workspaces": dschema.ListNestedAttribute{
				Computed:            true,
				NestedObject:        dschema.NestedAttributeObject{Attributes: workspaceNestedAttributes()},
				MarkdownDescription: "OpenRouter workspaces.",
			},
		},
	}
}

func workspaceNestedAttributes() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"id":                                  dschema.StringAttribute{Computed: true},
		"name":                                dschema.StringAttribute{Computed: true},
		"slug":                                dschema.StringAttribute{Computed: true},
		"description":                         dschema.StringAttribute{Computed: true},
		"default_text_model":                  dschema.StringAttribute{Computed: true},
		"default_image_model":                 dschema.StringAttribute{Computed: true},
		"default_provider_sort":               dschema.StringAttribute{Computed: true},
		"created_by":                          dschema.StringAttribute{Computed: true},
		"created_at":                          dschema.StringAttribute{Computed: true},
		"updated_at":                          dschema.StringAttribute{Computed: true},
		"is_data_discount_logging_enabled":    dschema.BoolAttribute{Computed: true},
		"is_observability_broadcast_enabled":  dschema.BoolAttribute{Computed: true},
		"is_observability_io_logging_enabled": dschema.BoolAttribute{Computed: true},
	}
}

func (d *workspacesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	workspaces, err := d.client.ListWorkspaces(ctx)
	if err != nil {
		resp.Diagnostics.AddError("List OpenRouter workspaces failed", err.Error())
		return
	}
	state := workspacesDataSourceModel{Workspaces: make([]workspaceDataModel, 0, len(workspaces))}
	for _, workspace := range workspaces {
		state.Workspaces = append(state.Workspaces, workspaceDataModel{
			ID:                              types.StringValue(workspace.ID),
			Name:                            types.StringValue(workspace.Name),
			Slug:                            types.StringValue(workspace.Slug),
			Description:                     stringValue(workspace.Description),
			DefaultTextModel:                stringValue(workspace.DefaultTextModel),
			DefaultImageModel:               stringValue(workspace.DefaultImageModel),
			DefaultProviderSort:             stringValue(workspace.DefaultProviderSort),
			CreatedBy:                       types.StringValue(workspace.CreatedBy),
			CreatedAt:                       types.StringValue(workspace.CreatedAt),
			UpdatedAt:                       types.StringValue(workspace.UpdatedAt),
			IsDataDiscountLoggingEnabled:    boolValue(workspace.IsDataDiscountLoggingEnabled),
			IsObservabilityBroadcastEnabled: boolValue(workspace.IsObservabilityBroadcastEnabled),
			IsObservabilityIOLoggingEnabled: boolValue(workspace.IsObservabilityIOLoggingEnabled),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
