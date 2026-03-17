package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGenerateDraft_MissingAPIKey(t *testing.T) {
	_, err := GenerateDraft("prompt", "model", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "ANTHROPIC_API_KEY") {
		t.Errorf("expected ANTHROPIC_API_KEY in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "export ANTHROPIC_API_KEY") {
		t.Errorf("expected export hint in error, got: %v", err)
	}
}

func TestGenerateDraft_UsesModel(t *testing.T) {
	wantModel := "claude-test-model"

	var gotModel string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req apiRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		gotModel = req.Model
		resp := apiResponse{
			Content: []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}{{Type: "text", Text: "## Context\nGenerated."}},
		}
		w.Header().Set("content-type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	// Temporarily override the URL by using the internal function.
	client := srv.Client()
	// Use a transport that rewrites the host.
	client.Transport = &rewriteTransport{base: http.DefaultTransport, target: srv.URL}

	_, err := generateDraft("prompt", wantModel, "test-key", client)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotModel != wantModel {
		t.Errorf("got model %q, want %q", gotModel, wantModel)
	}
}

func TestGenerateDraft_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := apiResponse{
			Error: &struct {
				Message string `json:"message"`
				Type    string `json:"type"`
			}{Message: "invalid api key", Type: "authentication_error"},
		}
		w.Header().Set("content-type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := srv.Client()
	client.Transport = &rewriteTransport{base: http.DefaultTransport, target: srv.URL}

	_, err := generateDraft("prompt", "model", "bad-key", client)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid api key") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDryRun_ShowsPromptAndConfirms(t *testing.T) {
	var out bytes.Buffer
	prompt := "Title: test\nChanged files:\n  foo.go\n"

	result := DryRun(prompt, strings.NewReader("y\n"), &out)
	if !result {
		t.Error("expected true for 'y' input")
	}
	got := out.String()
	if !strings.Contains(got, "--- Prompt to be sent to Anthropic API ---") {
		t.Errorf("missing header in output: %s", got)
	}
	if !strings.Contains(got, prompt) {
		t.Errorf("missing prompt in output: %s", got)
	}
	if !strings.Contains(got, "------------------------------------------") {
		t.Errorf("missing footer in output: %s", got)
	}
}

func TestDryRun_DeclinesOnN(t *testing.T) {
	var out bytes.Buffer
	result := DryRun("prompt", strings.NewReader("n\n"), &out)
	if result {
		t.Error("expected false for 'n' input")
	}
}

// rewriteTransport rewrites all requests to the given target URL (for testing).
type rewriteTransport struct {
	base   http.RoundTripper
	target string
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	newURL := fmt.Sprintf("%s%s", t.target, req.URL.Path)
	newReq := req.Clone(req.Context())
	parsed, err := newReq.URL.Parse(newURL)
	if err != nil {
		return nil, err
	}
	newReq.URL = parsed
	newReq.Host = parsed.Host
	return t.base.RoundTrip(newReq)
}
