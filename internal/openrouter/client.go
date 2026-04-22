package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultBaseURL   = "https://openrouter.ai/api/v1"
	DefaultUserAgent = "terraform-provider-openrouter/0.1.0"
	defaultPageLimit = 100
)

// Client is a small OpenRouter management API client tailored for the Terraform provider.
type Client struct {
	baseURL    *url.URL
	apiKey     string
	userAgent  string
	httpClient *http.Client
	maxRetries int
	minBackoff time.Duration
	maxBackoff time.Duration
	sleeper    func(context.Context, time.Duration) error
}

// Option customizes the API client.
type Option func(*Client)

// WithHTTPClient overrides the default HTTP client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

// WithRetryConfig overrides retry behavior. Primarily useful in tests.
func WithRetryConfig(maxRetries int, minBackoff, maxBackoff time.Duration) Option {
	return func(c *Client) {
		c.maxRetries = maxRetries
		c.minBackoff = minBackoff
		c.maxBackoff = maxBackoff
	}
}

// WithSleeper overrides retry sleeping. Primarily useful in tests.
func WithSleeper(sleeper func(context.Context, time.Duration) error) Option {
	return func(c *Client) {
		if sleeper != nil {
			c.sleeper = sleeper
		}
	}
}

// NewClient creates a new OpenRouter API client.
func NewClient(baseURL, apiKey, userAgent string, opts ...Option) (*Client, error) {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = DefaultBaseURL
	}
	parsed, err := url.Parse(strings.TrimRight(baseURL, "/"))
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("base url must be absolute")
	}

	ua := DefaultUserAgent
	if strings.TrimSpace(userAgent) != "" {
		ua = ua + " " + strings.TrimSpace(userAgent)
	}

	c := &Client{
		baseURL:    parsed,
		apiKey:     strings.TrimSpace(apiKey),
		userAgent:  ua,
		httpClient: http.DefaultClient,
		maxRetries: 4,
		minBackoff: 250 * time.Millisecond,
		maxBackoff: 4 * time.Second,
		sleeper: func(ctx context.Context, d time.Duration) error {
			timer := time.NewTimer(d)
			defer timer.Stop()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
				return nil
			}
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

// APIError is returned for non-successful API responses.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("openrouter API returned HTTP %d", e.StatusCode)
	}
	return fmt.Sprintf("openrouter API returned HTTP %d: %s", e.StatusCode, e.Body)
}

// IsNotFound reports whether err is a 404 API error.
func IsNotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

func (c *Client) newRequest(ctx context.Context, method, path string, query url.Values, body any) (*http.Request, error) {
	rel := &url.URL{Path: strings.TrimLeft(path, "/")}
	u := c.baseURL.ResolveReference(rel)
	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encode request body: %w", err)
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), reader)
	if err != nil {
		return nil, err
	}
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

func (c *Client) do(ctx context.Context, method, path string, query url.Values, body any, out any) error {
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		req, err := c.newRequest(ctx, method, path, query, body)
		if err != nil {
			return err
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if attempt == c.maxRetries {
				return err
			}
			if sleepErr := c.sleeper(ctx, c.backoff(attempt, nil)); sleepErr != nil {
				return sleepErr
			}
			continue
		}

		responseBody, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			return readErr
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if out == nil || len(responseBody) == 0 {
				return nil
			}
			if err := json.Unmarshal(responseBody, out); err != nil {
				return fmt.Errorf("decode response: %w", err)
			}
			return nil
		}

		apiErr := &APIError{StatusCode: resp.StatusCode, Body: strings.TrimSpace(string(responseBody))}
		lastErr = apiErr
		if !isRetryableStatus(resp.StatusCode) || attempt == c.maxRetries {
			return apiErr
		}
		if sleepErr := c.sleeper(ctx, c.backoff(attempt, resp)); sleepErr != nil {
			return sleepErr
		}
	}
	return lastErr
}

func isRetryableStatus(status int) bool {
	return status == http.StatusTooManyRequests || status == http.StatusBadGateway || status == http.StatusServiceUnavailable || status == http.StatusGatewayTimeout || status >= 500
}

func (c *Client) backoff(attempt int, resp *http.Response) time.Duration {
	if resp != nil {
		if retryAfter := parseRetryAfter(resp.Header.Get("Retry-After")); retryAfter > 0 {
			return retryAfter
		}
	}
	base := c.minBackoff
	if base <= 0 {
		base = time.Millisecond
	}
	d := base << attempt
	if c.maxBackoff > 0 && d > c.maxBackoff {
		d = c.maxBackoff
	}
	// Add small jitter so parallel Terraform operations do not retry in lockstep.
	if d > 0 {
		jitter := time.Duration(rand.Int63n(int64(d/4 + 1)))
		d += jitter
	}
	return d
}

func parseRetryAfter(value string) time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(value); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	if when, err := http.ParseTime(value); err == nil {
		return time.Until(when)
	}
	return 0
}

func pageQuery(offset, limit int, extra url.Values) url.Values {
	q := url.Values{}
	for key, values := range extra {
		for _, value := range values {
			q.Add(key, value)
		}
	}
	if offset > 0 {
		q.Set("offset", strconv.Itoa(offset))
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	return q
}

func addOptionalString(body map[string]any, key, value string) {
	if value != "" {
		body[key] = value
	}
}
