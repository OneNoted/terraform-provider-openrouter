package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/openrouter/terraform-provider-openrouter/internal/openrouter"
)

var (
	_ datasource.DataSource              = (*modelsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*modelsDataSource)(nil)
)

type modelsDataSource struct{ client *openrouter.Client }

type modelsDataSourceModel struct {
	Category types.String     `tfsdk:"category"`
	Models   []modelItemModel `tfsdk:"models"`
}

type modelItemModel struct {
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	Created               types.Int64  `tfsdk:"created"`
	Description           types.String `tfsdk:"description"`
	CanonicalSlug         types.String `tfsdk:"canonical_slug"`
	ContextLength         types.Int64  `tfsdk:"context_length"`
	HuggingFaceID         types.String `tfsdk:"hugging_face_id"`
	InputModalities       types.List   `tfsdk:"input_modalities"`
	OutputModalities      types.List   `tfsdk:"output_modalities"`
	Tokenizer             types.String `tfsdk:"tokenizer"`
	InstructType          types.String `tfsdk:"instruct_type"`
	TopProviderModerated  types.Bool   `tfsdk:"top_provider_is_moderated"`
	TopProviderContext    types.Int64  `tfsdk:"top_provider_context_length"`
	MaxCompletionTokens   types.Int64  `tfsdk:"top_provider_max_completion_tokens"`
	PricingPrompt         types.String `tfsdk:"pricing_prompt"`
	PricingCompletion     types.String `tfsdk:"pricing_completion"`
	PricingImage          types.String `tfsdk:"pricing_image"`
	PricingRequest        types.String `tfsdk:"pricing_request"`
	PricingWebSearch      types.String `tfsdk:"pricing_web_search"`
	PricingReasoning      types.String `tfsdk:"pricing_internal_reasoning"`
	PricingCacheRead      types.String `tfsdk:"pricing_input_cache_read"`
	PricingCacheWrite     types.String `tfsdk:"pricing_input_cache_write"`
	SupportedParameters   types.List   `tfsdk:"supported_parameters"`
	DefaultParametersJSON types.String `tfsdk:"default_parameters_json"`
	PerRequestLimitsJSON  types.String `tfsdk:"per_request_limits_json"`
}

func NewModelsDataSource() datasource.DataSource { return &modelsDataSource{} }

func (d *modelsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_models"
}

func (d *modelsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *modelsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		MarkdownDescription: "Lists OpenRouter model metadata.",
		Attributes: map[string]dschema.Attribute{
			"category": dschema.StringAttribute{Optional: true, MarkdownDescription: "Optional OpenRouter model category filter, such as `programming`."},
			"models": dschema.ListNestedAttribute{
				Computed:            true,
				NestedObject:        dschema.NestedAttributeObject{Attributes: modelNestedAttributes()},
				MarkdownDescription: "OpenRouter models.",
			},
		},
	}
}

func modelNestedAttributes() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"id":                                 dschema.StringAttribute{Computed: true},
		"name":                               dschema.StringAttribute{Computed: true},
		"created":                            dschema.Int64Attribute{Computed: true},
		"description":                        dschema.StringAttribute{Computed: true},
		"canonical_slug":                     dschema.StringAttribute{Computed: true},
		"context_length":                     dschema.Int64Attribute{Computed: true},
		"hugging_face_id":                    dschema.StringAttribute{Computed: true},
		"input_modalities":                   dschema.ListAttribute{Computed: true, ElementType: types.StringType},
		"output_modalities":                  dschema.ListAttribute{Computed: true, ElementType: types.StringType},
		"tokenizer":                          dschema.StringAttribute{Computed: true},
		"instruct_type":                      dschema.StringAttribute{Computed: true},
		"top_provider_is_moderated":          dschema.BoolAttribute{Computed: true},
		"top_provider_context_length":        dschema.Int64Attribute{Computed: true},
		"top_provider_max_completion_tokens": dschema.Int64Attribute{Computed: true},
		"pricing_prompt":                     dschema.StringAttribute{Computed: true},
		"pricing_completion":                 dschema.StringAttribute{Computed: true},
		"pricing_image":                      dschema.StringAttribute{Computed: true},
		"pricing_request":                    dschema.StringAttribute{Computed: true},
		"pricing_web_search":                 dschema.StringAttribute{Computed: true},
		"pricing_internal_reasoning":         dschema.StringAttribute{Computed: true},
		"pricing_input_cache_read":           dschema.StringAttribute{Computed: true},
		"pricing_input_cache_write":          dschema.StringAttribute{Computed: true},
		"supported_parameters":               dschema.ListAttribute{Computed: true, ElementType: types.StringType},
		"default_parameters_json":            dschema.StringAttribute{Computed: true},
		"per_request_limits_json":            dschema.StringAttribute{Computed: true},
	}
}

func (d *modelsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		addProviderNotConfiguredError(&resp.Diagnostics)
		return
	}
	var config modelsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	category := ""
	if !config.Category.IsNull() && !config.Category.IsUnknown() {
		category = config.Category.ValueString()
	}
	models, err := d.client.ListModels(ctx, category)
	if err != nil {
		resp.Diagnostics.AddError("List OpenRouter models failed", err.Error())
		return
	}
	state := modelsDataSourceModel{Category: config.Category, Models: make([]modelItemModel, 0, len(models))}
	for _, model := range models {
		inputModalities, diags := stringListValue(ctx, model.Architecture.InputModalities)
		resp.Diagnostics.Append(diags...)
		outputModalities, diags := stringListValue(ctx, model.Architecture.OutputModalities)
		resp.Diagnostics.Append(diags...)
		supportedParameters, diags := stringListValue(ctx, model.SupportedParameters)
		resp.Diagnostics.Append(diags...)
		state.Models = append(state.Models, modelItemModel{
			ID:                    types.StringValue(model.ID),
			Name:                  types.StringValue(model.Name),
			Created:               types.Int64Value(model.Created),
			Description:           types.StringValue(model.Description),
			CanonicalSlug:         types.StringValue(model.CanonicalSlug),
			ContextLength:         int64Value(model.ContextLength),
			HuggingFaceID:         stringValue(model.HuggingFaceID),
			InputModalities:       inputModalities,
			OutputModalities:      outputModalities,
			Tokenizer:             types.StringValue(model.Architecture.Tokenizer),
			InstructType:          stringValue(model.Architecture.InstructType),
			TopProviderModerated:  boolValue(model.TopProvider.IsModerated),
			TopProviderContext:    int64Value(model.TopProvider.ContextLength),
			MaxCompletionTokens:   int64Value(model.TopProvider.MaxCompletionTokens),
			PricingPrompt:         types.StringValue(string(model.Pricing.Prompt)),
			PricingCompletion:     types.StringValue(string(model.Pricing.Completion)),
			PricingImage:          types.StringValue(string(model.Pricing.Image)),
			PricingRequest:        types.StringValue(string(model.Pricing.Request)),
			PricingWebSearch:      types.StringValue(string(model.Pricing.WebSearch)),
			PricingReasoning:      types.StringValue(string(model.Pricing.InternalReasoning)),
			PricingCacheRead:      types.StringValue(string(model.Pricing.InputCacheRead)),
			PricingCacheWrite:     types.StringValue(string(model.Pricing.InputCacheWrite)),
			SupportedParameters:   supportedParameters,
			DefaultParametersJSON: jsonString(model.DefaultParameters),
			PerRequestLimitsJSON:  jsonString(model.PerRequestLimits),
		})
	}
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func int64Value(value *int64) types.Int64 {
	if value == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*value)
}
