package openrouter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestClientRejectsPlainHTTPForNonLocalBaseURL(t *testing.T) {
	if _, err := NewClient("http://openrouter.example/api/v1", "management-key", ""); err == nil {
		t.Fatal("expected non-local plaintext HTTP base URL to be rejected")
	}
}

func TestClientDoesNotRetryPostCreate(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"transient"}`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "management-key", "", WithRetryConfig(3, time.Millisecond, time.Millisecond), WithSleeper(func(ctx context.Context, d time.Duration) error {
		return nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.CreateAPIKey(context.Background(), map[string]any{"name": "ci-key"}); err == nil {
		t.Fatal("expected create error")
	}
	if calls != 1 {
		t.Fatalf("POST calls = %d, want 1", calls)
	}
}

func TestClientSendsAuthAndUserAgent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workspaces" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if got, want := r.Header.Get("Authorization"), "Bearer management-key"; got != want {
			t.Fatalf("Authorization = %q, want %q", got, want)
		}
		if got := r.Header.Get("User-Agent"); !strings.Contains(got, DefaultUserAgent) || !strings.Contains(got, "custom-agent") {
			t.Fatalf("User-Agent = %q, want default and custom suffix", got)
		}
		_, _ = w.Write([]byte(`{"data":[],"total_count":0}`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "management-key", "custom-agent")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListWorkspaces(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestClientPreservesBaseURLPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/api/v1/workspaces"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		_, _ = w.Write([]byte(`{"data":[],"total_count":0}`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL+"/api/v1", "management-key", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListWorkspaces(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestClientRetries429WithRetryAfter(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"slow down"}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":[],"total_count":0}`))
	}))
	defer server.Close()

	var slept []time.Duration
	client, err := NewClient(server.URL, "management-key", "", WithRetryConfig(2, time.Millisecond, time.Millisecond), WithSleeper(func(ctx context.Context, d time.Duration) error {
		slept = append(slept, d)
		return nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListWorkspaces(context.Background()); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
	if len(slept) != 1 || slept[0] != time.Second {
		t.Fatalf("slept = %v, want [1s]", slept)
	}
}

func TestClientListWorkspacesPaginates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		offset := r.URL.Query().Get("offset")
		limit := r.URL.Query().Get("limit")
		if limit != "100" {
			t.Fatalf("limit = %q, want 100", limit)
		}
		items := make([]Workspace, 0, 100)
		switch offset {
		case "":
			for i := 0; i < 100; i++ {
				items = append(items, Workspace{ID: "workspace-a", Name: "A", Slug: "a"})
			}
		case "100":
			items = append(items, Workspace{ID: "workspace-b", Name: "B", Slug: "b"})
		default:
			t.Fatalf("unexpected offset %q", offset)
		}
		_ = json.NewEncoder(w).Encode(workspaceListResponse{Data: items, TotalCount: 101})
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "management-key", "")
	if err != nil {
		t.Fatal(err)
	}
	workspaces, err := client.ListWorkspaces(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(workspaces) != 101 {
		t.Fatalf("len(workspaces) = %d, want 101", len(workspaces))
	}
}

func TestClientCreateAPIKeyDiscardsRawKeyMaterial(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/keys" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"data":{"hash":"key-hash","name":"ci-key","label":"ci-key"},"key":"sk-or-v1-plaintext-secret"}`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "management-key", "")
	if err != nil {
		t.Fatal(err)
	}
	apiKey, err := client.CreateAPIKey(context.Background(), map[string]any{"name": "ci-key"})
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(apiKey)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(payload), "plaintext-secret") || strings.Contains(string(payload), "\"key\"") {
		t.Fatalf("API key metadata leaked raw key material: %s", payload)
	}
	if apiKey.Hash != "key-hash" {
		t.Fatalf("Hash = %q, want key-hash", apiKey.Hash)
	}
}

func TestIsNotFound(t *testing.T) {
	if !IsNotFound(&APIError{StatusCode: http.StatusNotFound}) {
		t.Fatal("expected IsNotFound to match 404 APIError")
	}
	if IsNotFound(&APIError{StatusCode: http.StatusInternalServerError}) {
		t.Fatal("expected IsNotFound not to match 500 APIError")
	}
}

func TestStringNumberAcceptsStringAndNumberJSON(t *testing.T) {
	var payload struct {
		StringValue StringNumber `json:"string_value"`
		NumberValue StringNumber `json:"number_value"`
	}
	if err := json.Unmarshal([]byte(`{"string_value":"0.00003","number_value":0.00006}`), &payload); err != nil {
		t.Fatal(err)
	}
	if got, want := string(payload.StringValue), "0.00003"; got != want {
		t.Fatalf("StringValue = %q, want %q", got, want)
	}
	if got, want := string(payload.NumberValue), "0.00006"; got != want {
		t.Fatalf("NumberValue = %q, want %q", got, want)
	}
}
