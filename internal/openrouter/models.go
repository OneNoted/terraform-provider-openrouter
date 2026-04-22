package openrouter

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type workspaceResponse struct {
	Data Workspace `json:"data"`
}

type workspaceListResponse struct {
	Data       []Workspace `json:"data"`
	TotalCount int         `json:"total_count"`
}

// Workspace is an OpenRouter workspace.
type Workspace struct {
	ID                              string  `json:"id"`
	Name                            string  `json:"name"`
	Slug                            string  `json:"slug"`
	Description                     *string `json:"description"`
	DefaultTextModel                *string `json:"default_text_model"`
	DefaultImageModel               *string `json:"default_image_model"`
	DefaultProviderSort             *string `json:"default_provider_sort"`
	CreatedBy                       string  `json:"created_by"`
	CreatedAt                       string  `json:"created_at"`
	UpdatedAt                       string  `json:"updated_at"`
	IsDataDiscountLoggingEnabled    *bool   `json:"is_data_discount_logging_enabled"`
	IsObservabilityBroadcastEnabled *bool   `json:"is_observability_broadcast_enabled"`
	IsObservabilityIOLoggingEnabled *bool   `json:"is_observability_io_logging_enabled"`
}

type DeleteResponse struct {
	Deleted bool `json:"deleted"`
	Data    *struct {
		Success bool `json:"success"`
	} `json:"data,omitempty"`
}

func (c *Client) CreateWorkspace(ctx context.Context, body map[string]any) (*Workspace, error) {
	var response workspaceResponse
	if err := c.do(ctx, http.MethodPost, "/workspaces", nil, body, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) GetWorkspace(ctx context.Context, id string) (*Workspace, error) {
	var response workspaceResponse
	if err := c.do(ctx, http.MethodGet, "/workspaces/"+url.PathEscape(id), nil, nil, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) UpdateWorkspace(ctx context.Context, id string, body map[string]any) (*Workspace, error) {
	var response workspaceResponse
	if err := c.do(ctx, http.MethodPatch, "/workspaces/"+url.PathEscape(id), nil, body, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) DeleteWorkspace(ctx context.Context, id string) error {
	var response DeleteResponse
	return c.do(ctx, http.MethodDelete, "/workspaces/"+url.PathEscape(id), nil, nil, &response)
}

func (c *Client) ListWorkspaces(ctx context.Context) ([]Workspace, error) {
	var all []Workspace
	for offset := 0; ; offset += defaultPageLimit {
		var response workspaceListResponse
		if err := c.do(ctx, http.MethodGet, "/workspaces", pageQuery(offset, defaultPageLimit, nil), nil, &response); err != nil {
			return nil, err
		}
		all = append(all, response.Data...)
		if len(response.Data) < defaultPageLimit || (response.TotalCount > 0 && len(all) >= response.TotalCount) {
			return all, nil
		}
	}
}

type apiKeyResponse struct {
	Data APIKey `json:"data"`
	Key  string `json:"key"`
}

type apiKeyListResponse struct {
	Data []APIKey `json:"data"`
}

// APIKey is stable API key metadata. It intentionally omits create-time raw key material.
type APIKey struct {
	Hash               string   `json:"hash"`
	Name               string   `json:"name"`
	Label              string   `json:"label"`
	WorkspaceID        *string  `json:"workspace_id"`
	Disabled           *bool    `json:"disabled"`
	Limit              *float64 `json:"limit"`
	LimitRemaining     *float64 `json:"limit_remaining"`
	LimitReset         *string  `json:"limit_reset"`
	IncludeBYOKInLimit *bool    `json:"include_byok_in_limit"`
	Usage              *float64 `json:"usage"`
	UsageDaily         *float64 `json:"usage_daily"`
	UsageWeekly        *float64 `json:"usage_weekly"`
	UsageMonthly       *float64 `json:"usage_monthly"`
	BYOKUsage          *float64 `json:"byok_usage"`
	BYOKUsageDaily     *float64 `json:"byok_usage_daily"`
	BYOKUsageWeekly    *float64 `json:"byok_usage_weekly"`
	BYOKUsageMonthly   *float64 `json:"byok_usage_monthly"`
	CreatedAt          string   `json:"created_at"`
	UpdatedAt          string   `json:"updated_at"`
	ExpiresAt          *string  `json:"expires_at"`
	CreatorUserID      string   `json:"creator_user_id"`
}

func (c *Client) CreateAPIKey(ctx context.Context, body map[string]any) (*APIKey, error) {
	var response apiKeyResponse
	if err := c.do(ctx, http.MethodPost, "/keys", nil, body, &response); err != nil {
		return nil, err
	}
	// response.Key contains the one-time raw key. Do not return or persist it.
	return &response.Data, nil
}

func (c *Client) GetAPIKey(ctx context.Context, hash string) (*APIKey, error) {
	var response apiKeyResponse
	if err := c.do(ctx, http.MethodGet, "/keys/"+url.PathEscape(hash), nil, nil, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) UpdateAPIKey(ctx context.Context, hash string, body map[string]any) (*APIKey, error) {
	var response apiKeyResponse
	if err := c.do(ctx, http.MethodPatch, "/keys/"+url.PathEscape(hash), nil, body, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) DeleteAPIKey(ctx context.Context, hash string) error {
	var response DeleteResponse
	return c.do(ctx, http.MethodDelete, "/keys/"+url.PathEscape(hash), nil, nil, &response)
}

func (c *Client) ListAPIKeys(ctx context.Context, includeDisabled bool) ([]APIKey, error) {
	var all []APIKey
	for offset := 0; ; offset += defaultPageLimit {
		q := url.Values{}
		q.Set("include_disabled", fmt.Sprintf("%t", includeDisabled))
		q.Set("offset", fmt.Sprintf("%d", offset))
		var response apiKeyListResponse
		if err := c.do(ctx, http.MethodGet, "/keys", q, nil, &response); err != nil {
			return nil, err
		}
		all = append(all, response.Data...)
		if len(response.Data) < defaultPageLimit {
			return all, nil
		}
	}
}

type guardrailResponse struct {
	Data Guardrail `json:"data"`
}

type guardrailListResponse struct {
	Data       []Guardrail `json:"data"`
	TotalCount int         `json:"total_count"`
}

// Guardrail is an OpenRouter guardrail.
type Guardrail struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	WorkspaceID      *string  `json:"workspace_id"`
	Description      *string  `json:"description"`
	AllowedModels    []string `json:"allowed_models"`
	AllowedProviders []string `json:"allowed_providers"`
	IgnoredModels    []string `json:"ignored_models"`
	IgnoredProviders []string `json:"ignored_providers"`
	EnforceZDR       *bool    `json:"enforce_zdr"`
	LimitUSD         *float64 `json:"limit_usd"`
	ResetInterval    *string  `json:"reset_interval"`
	CreatedAt        string   `json:"created_at"`
	UpdatedAt        string   `json:"updated_at"`
}

func (c *Client) CreateGuardrail(ctx context.Context, body map[string]any) (*Guardrail, error) {
	var response guardrailResponse
	if err := c.do(ctx, http.MethodPost, "/guardrails", nil, body, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) GetGuardrail(ctx context.Context, id string) (*Guardrail, error) {
	var response guardrailResponse
	if err := c.do(ctx, http.MethodGet, "/guardrails/"+url.PathEscape(id), nil, nil, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) UpdateGuardrail(ctx context.Context, id string, body map[string]any) (*Guardrail, error) {
	var response guardrailResponse
	if err := c.do(ctx, http.MethodPatch, "/guardrails/"+url.PathEscape(id), nil, body, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) DeleteGuardrail(ctx context.Context, id string) error {
	var response DeleteResponse
	return c.do(ctx, http.MethodDelete, "/guardrails/"+url.PathEscape(id), nil, nil, &response)
}

func (c *Client) ListGuardrails(ctx context.Context, workspaceID string) ([]Guardrail, error) {
	var all []Guardrail
	extra := url.Values{}
	if workspaceID != "" {
		extra.Set("workspace_id", workspaceID)
	}
	for offset := 0; ; offset += defaultPageLimit {
		var response guardrailListResponse
		if err := c.do(ctx, http.MethodGet, "/guardrails", pageQuery(offset, defaultPageLimit, extra), nil, &response); err != nil {
			return nil, err
		}
		all = append(all, response.Data...)
		if len(response.Data) < defaultPageLimit || (response.TotalCount > 0 && len(all) >= response.TotalCount) {
			return all, nil
		}
	}
}

type providersResponse struct {
	Data []Provider `json:"data"`
}

// Provider is OpenRouter provider metadata.
type Provider struct {
	Name              string   `json:"name"`
	Slug              string   `json:"slug"`
	PrivacyPolicyURL  string   `json:"privacy_policy_url"`
	TermsOfServiceURL string   `json:"terms_of_service_url"`
	StatusPageURL     string   `json:"status_page_url"`
	Headquarters      string   `json:"headquarters"`
	Datacenters       []string `json:"datacenters"`
}

func (c *Client) ListProviders(ctx context.Context) ([]Provider, error) {
	var response providersResponse
	if err := c.do(ctx, http.MethodGet, "/providers", nil, nil, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}

type modelsResponse struct {
	Data []Model `json:"data"`
}

// Model is OpenRouter model metadata.
type Model struct {
	ID                  string            `json:"id"`
	Name                string            `json:"name"`
	Created             int64             `json:"created"`
	Description         string            `json:"description"`
	CanonicalSlug       string            `json:"canonical_slug"`
	ContextLength       *int64            `json:"context_length"`
	HuggingFaceID       string            `json:"hugging_face_id"`
	Architecture        ModelArchitecture `json:"architecture"`
	TopProvider         ModelTopProvider  `json:"top_provider"`
	Pricing             ModelPricing      `json:"pricing"`
	SupportedParameters []string          `json:"supported_parameters"`
	DefaultParameters   map[string]any    `json:"default_parameters"`
	PerRequestLimits    map[string]any    `json:"per_request_limits"`
}

type ModelArchitecture struct {
	InputModalities  []string `json:"input_modalities"`
	OutputModalities []string `json:"output_modalities"`
	Tokenizer        string   `json:"tokenizer"`
	InstructType     string   `json:"instruct_type"`
}

type ModelTopProvider struct {
	IsModerated         *bool  `json:"is_moderated"`
	ContextLength       *int64 `json:"context_length"`
	MaxCompletionTokens *int64 `json:"max_completion_tokens"`
}

type ModelPricing struct {
	Prompt            string `json:"prompt"`
	Completion        string `json:"completion"`
	Image             string `json:"image"`
	Request           string `json:"request"`
	WebSearch         string `json:"web_search"`
	InternalReasoning string `json:"internal_reasoning"`
	InputCacheRead    string `json:"input_cache_read"`
	InputCacheWrite   string `json:"input_cache_write"`
}

func (c *Client) ListModels(ctx context.Context, category string) ([]Model, error) {
	q := url.Values{}
	if category != "" {
		q.Set("category", category)
	}
	var response modelsResponse
	if err := c.do(ctx, http.MethodGet, "/models", q, nil, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}
