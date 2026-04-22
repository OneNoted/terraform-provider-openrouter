package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/openrouter/terraform-provider-openrouter/internal/openrouter"
)

var (
	_ datasource.DataSource              = (*providersDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*providersDataSource)(nil)
)

type providersDataSource struct{ client *openrouter.Client }

type providersDataSourceModel struct {
	Providers []providerItemModel `tfsdk:"providers"`
}

type providerItemModel struct {
	Name              types.String `tfsdk:"name"`
	Slug              types.String `tfsdk:"slug"`
	PrivacyPolicyURL  types.String `tfsdk:"privacy_policy_url"`
	TermsOfServiceURL types.String `tfsdk:"terms_of_service_url"`
	StatusPageURL     types.String `tfsdk:"status_page_url"`
	Headquarters      types.String `tfsdk:"headquarters"`
	Datacenters       types.List   `tfsdk:"datacenters"`
}

func NewProvidersDataSource() datasource.DataSource { return &providersDataSource{} }

func (d *providersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_providers"
}

func (d *providersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *providersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		MarkdownDescription: "Lists OpenRouter provider metadata.",
		Attributes: map[string]dschema.Attribute{
			"providers": dschema.ListNestedAttribute{
				Computed:            true,
				NestedObject:        dschema.NestedAttributeObject{Attributes: providerNestedAttributes()},
				MarkdownDescription: "OpenRouter providers.",
			},
		},
	}
}

func providerNestedAttributes() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"name":                 dschema.StringAttribute{Computed: true},
		"slug":                 dschema.StringAttribute{Computed: true},
		"privacy_policy_url":   dschema.StringAttribute{Computed: true},
		"terms_of_service_url": dschema.StringAttribute{Computed: true},
		"status_page_url":      dschema.StringAttribute{Computed: true},
		"headquarters":         dschema.StringAttribute{Computed: true},
		"datacenters":          dschema.ListAttribute{Computed: true, ElementType: types.StringType},
	}
}

func (d *providersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		addProviderNotConfiguredError(&resp.Diagnostics)
		return
	}
	providers, err := d.client.ListProviders(ctx)
	if err != nil {
		resp.Diagnostics.AddError("List OpenRouter providers failed", err.Error())
		return
	}
	state := providersDataSourceModel{Providers: make([]providerItemModel, 0, len(providers))}
	for _, provider := range providers {
		datacenters, diags := stringListValue(ctx, provider.Datacenters)
		resp.Diagnostics.Append(diags...)
		state.Providers = append(state.Providers, providerItemModel{
			Name:              types.StringValue(provider.Name),
			Slug:              types.StringValue(provider.Slug),
			PrivacyPolicyURL:  stringValue(provider.PrivacyPolicyURL),
			TermsOfServiceURL: stringValue(provider.TermsOfServiceURL),
			StatusPageURL:     stringValue(provider.StatusPageURL),
			Headquarters:      stringValue(provider.Headquarters),
			Datacenters:       datacenters,
		})
	}
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
