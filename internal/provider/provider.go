package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	providerschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/openrouter/terraform-provider-openrouter/internal/openrouter"
)

var _ provider.Provider = (*openRouterProvider)(nil)

type openRouterProvider struct {
	version string
}

type providerModel struct {
	ManagementAPIKey types.String `tfsdk:"management_api_key"`
	BaseURL          types.String `tfsdk:"base_url"`
	UserAgent        types.String `tfsdk:"user_agent"`
}

// New returns a Terraform Plugin Framework provider factory.
func New() provider.Provider {
	return &openRouterProvider{version: "0.1.0"}
}

func (p *openRouterProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "openrouter"
	resp.Version = p.version
}

func (p *openRouterProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = providerschema.Schema{
		MarkdownDescription: "Manage OpenRouter workspaces, API keys, guardrails, and read model/provider metadata.",
		Attributes: map[string]providerschema.Attribute{
			"management_api_key": providerschema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "OpenRouter Management API key. May also be set with `OPENROUTER_MANAGEMENT_API_KEY` or `OPENROUTER_API_KEY`.",
			},
			"base_url": providerschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: fmt.Sprintf("OpenRouter API base URL. Defaults to `%s`.", openrouter.DefaultBaseURL),
			},
			"user_agent": providerschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional user-agent suffix appended to the provider default.",
			},
		},
	}
}

func (p *openRouterProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := os.Getenv("OPENROUTER_MANAGEMENT_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("OPENROUTER_API_KEY")
	}
	if !config.ManagementAPIKey.IsNull() && !config.ManagementAPIKey.IsUnknown() {
		apiKey = config.ManagementAPIKey.ValueString()
	}
	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("management_api_key"),
			"Missing OpenRouter Management API Key",
			"Set management_api_key in provider configuration or OPENROUTER_MANAGEMENT_API_KEY/OPENROUTER_API_KEY in the environment.",
		)
		return
	}

	baseURL := openrouter.DefaultBaseURL
	if !config.BaseURL.IsNull() && !config.BaseURL.IsUnknown() && config.BaseURL.ValueString() != "" {
		baseURL = config.BaseURL.ValueString()
	}
	userAgent := ""
	if !config.UserAgent.IsNull() && !config.UserAgent.IsUnknown() {
		userAgent = config.UserAgent.ValueString()
	}

	client, err := openrouter.NewClient(baseURL, apiKey, userAgent)
	if err != nil {
		resp.Diagnostics.AddError("Invalid OpenRouter client configuration", err.Error())
		return
	}
	tflog.Debug(ctx, "configured OpenRouter provider client", map[string]any{"base_url": baseURL})
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *openRouterProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewWorkspaceResource,
		NewAPIKeyResource,
		NewGuardrailResource,
	}
}

func (p *openRouterProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewProvidersDataSource,
		NewModelsDataSource,
		NewWorkspacesDataSource,
	}
}
