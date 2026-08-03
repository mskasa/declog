package cmd

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	kizamicontext "github.com/mskasa/kizami/internal/context"
)

// connectTestMCPSession wires up an in-process client/server pair over an in-memory
// transport (no real stdio/subprocess) so tool calls can be exercised end-to-end,
// including the SDK's own schema validation and JSON-RPC dispatch.
func connectTestMCPSession(t *testing.T, root string, dirs []string) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()

	server := newMCPServer(root, dirs)
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0"}, nil)

	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	if _, err := server.Connect(ctx, serverTransport, nil); err != nil {
		t.Fatalf("server.Connect: %v", err)
	}
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client.Connect: %v", err)
	}
	t.Cleanup(func() { session.Close() })
	return session
}

func callToolJSON(t *testing.T, session *mcp.ClientSession, name string, args map[string]any, out any) *mcp.CallToolResult {
	t.Helper()
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      name,
		Arguments: args,
	})
	if err != nil {
		t.Fatalf("CallTool(%s): %v", name, err)
	}
	if result.IsError {
		t.Fatalf("CallTool(%s) returned a tool error: %+v", name, result.Content)
	}
	if len(result.Content) == 0 {
		t.Fatalf("CallTool(%s): expected content, got none", name)
	}
	text, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("CallTool(%s): expected TextContent, got %T", name, result.Content[0])
	}
	if out != nil {
		if err := json.Unmarshal([]byte(text.Text), out); err != nil {
			t.Fatalf("CallTool(%s): unmarshaling result: %v\ncontent: %s", name, err, text.Text)
		}
	}
	return result
}

func TestMCP_DecisionsForFiles(t *testing.T) {
	root := newTestRepo(t)
	dir := decisionsPath(root)
	path := seedDecision(t, dir, 1, "Use Go", "Active")
	appendRelatedFile(t, path, "internal/search/search.go")
	writeRelatedFile(t, root, "internal/search/search.go")

	session := connectTestMCPSession(t, root, []string{dir})

	var out kizamicontext.Result
	callToolJSON(t, session, "kizami_decisions_for_files", map[string]any{
		"paths": []string{"internal/search/search.go"},
	}, &out)

	if len(out.Decisions) != 1 {
		t.Fatalf("expected 1 decision, got %d", len(out.Decisions))
	}
	if out.Decisions[0].Slug != "use-go" {
		t.Errorf("unexpected slug: %s", out.Decisions[0].Slug)
	}
	if out.Decisions[0].Drift.State != "ok" {
		t.Errorf("expected drift ok, got %s", out.Decisions[0].Drift.State)
	}
}

func TestMCP_DecisionsForFiles_RejectsEmptyPaths(t *testing.T) {
	root := newTestRepo(t)
	dir := decisionsPath(root)
	seedDecision(t, dir, 1, "Use Go", "Active")

	session := connectTestMCPSession(t, root, []string{dir})

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "kizami_decisions_for_files",
		Arguments: map[string]any{"paths": []string{}},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if !result.IsError {
		t.Error("expected a tool error for empty paths")
	}
}

func TestMCP_SearchDecisions(t *testing.T) {
	root := newTestRepo(t)
	dir := decisionsPath(root)
	seedDecision(t, dir, 1, "Use PostgreSQL over SQLite", "Active")

	session := connectTestMCPSession(t, root, []string{dir})

	var out searchDecisionsOutput
	callToolJSON(t, session, "kizami_search_decisions", map[string]any{
		"query": "postgresql",
	}, &out)

	if len(out.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(out.Results))
	}
	if out.Results[0].Slug != "use-postgresql-over-sqlite" {
		t.Errorf("unexpected slug: %s", out.Results[0].Slug)
	}
}

func TestMCP_GetDecision(t *testing.T) {
	root := newTestRepo(t)
	dir := decisionsPath(root)
	seedDecision(t, dir, 1, "Use Go", "Active")

	session := connectTestMCPSession(t, root, []string{dir})

	var out getDecisionOutput
	callToolJSON(t, session, "kizami_get_decision", map[string]any{
		"slug": "use-go",
	}, &out)

	if len(out.Matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(out.Matches))
	}
	if out.Matches[0].Markdown == "" {
		t.Error("expected non-empty markdown content")
	}
}

func TestMCP_GetDecision_NotFound(t *testing.T) {
	root := newTestRepo(t)
	dir := decisionsPath(root)
	seedDecision(t, dir, 1, "Use Go", "Active")

	session := connectTestMCPSession(t, root, []string{dir})

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "kizami_get_decision",
		Arguments: map[string]any{"slug": "does-not-exist"},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if !result.IsError {
		t.Error("expected a tool error for a missing slug")
	}
}
