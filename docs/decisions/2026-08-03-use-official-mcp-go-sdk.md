# Use the Official MCP Go SDK for `kizami mcp`

- Date: 2026-08-03
- Type: ADR
- Status: Active
- Author: masahiro.kasatani

## Context

Step 3 of the Agent Context Layer ([[agent-context-layer-design]]) needs `kizami mcp`, an MCP server exposing the `internal/context` resolver over stdio. The design doc left the transport/SDK choice as an open question: either the official `github.com/modelcontextprotocol/go-sdk`, or a small hand-rolled stdio JSON-RPC implementation, consistent with `use-go-over-shell-script`'s general preference for avoiding dependencies where the stdlib-adjacent effort is modest.

Checking the actual state of the SDK before deciding (rather than assuming): `github.com/modelcontextprotocol/go-sdk` reached v1.0.0 with an explicit API compatibility guarantee, is maintained in collaboration with Google, and is at v1.7.0 as of this writing, supporting multiple MCP protocol spec versions (2024-11-05 through 2026-07-28). It provides `mcp.NewServer`, a generic `mcp.AddTool[In, Out]` that derives JSON Schema from Go struct tags, and a ready-made `mcp.StdioTransport`.

Adding it is not free: `go get` pulls in the SDK plus `github.com/google/jsonschema-go` and several transitive dependencies (`segmentio/encoding`, `golang.org/x/oauth2`, `golang.org/x/sync`, `golang.org/x/time`, `yosida95/uritemplate`), and the SDK requires Go ≥ 1.25, which raises this module's `go` directive from 1.24. That's a real cost for a project whose stated tech-stack philosophy favors minimal dependencies — worth a deliberate decision, not a default.

## Decision

Use the official SDK. The deciding factor isn't the code size (a minimal JSON-RPC-over-stdio implementation would indeed be a few hundred lines) but protocol surface: MCP's handshake, capability negotiation, and JSON-RPC framing are a moving target across spec versions, and the SDK's job is exactly to absorb that churn so `kizami mcp`'s own code stays about the domain logic (resolving decisions), not about tracking protocol revisions. This is the same reasoning that led to choosing cobra over hand-rolled CLI parsing (`use-cobra-for-cli-framework`): a de facto standard maintained by the protocol's own ecosystem is worth taking a dependency for, when the alternative is re-implementing (and re-maintaining) a spec that isn't ours to define.

The Go version floor moves to 1.25 as a consequence, accepted knowingly rather than avoided by staying on a hand-rolled implementation. `.mise.toml`, `CONTRIBUTING.md`, and this repository's CLAUDE.md tech-stack line are updated to match; CI and the release workflow already read the Go version from `go.mod` (`go-version-file: go.mod`) so they adapt automatically.

## Consequences

- `kizami mcp` benefits from upstream fixes and future MCP spec support without kizami needing to track the protocol itself.
- The dependency tree grows by 7 modules (SDK + `jsonschema-go` + 5 further transitive deps), all `// indirect` except the SDK and `jsonschema-go` if used directly for schema tags.
- Contributors and `go install` users now need Go ≥ 1.25 (or rely on Go's toolchain auto-download, which most setups already have via `go.mod`'s `go` directive).
- Tool input/output schemas are declared as Go structs with `json`/`jsonschema` tags (see [[mcp-tools-as-questions-not-verbs]]), not hand-written JSON Schema documents.

## Alternatives Considered

**Hand-rolled stdio JSON-RPC 2.0**
No new dependency, no Go version bump. But MCP is not just JSON-RPC framing — it includes an initialization handshake, capability negotiation, and a schema format for tool definitions, all of which would need to be implemented and kept current against future spec revisions by hand. Given the SDK is official, stable (v1.0+ compatibility guarantee), and actively maintained, re-implementing this ourselves would mean carrying protocol-maintenance cost indefinitely for no corresponding benefit to kizami's actual value proposition (which is the decision data and resolution logic, not the transport).

**Wait for a lower-dependency or pre-1.0 alternative**
There isn't a materially lighter official alternative; unofficial community SDKs would carry more risk (less scrutiny, no compatibility guarantee) for the same or worse dependency cost.

## Related Files

- `go.mod`
- `internal/mcp/`
- `cmd/mcp.go`
- `.mise.toml`
- `CONTRIBUTING.md`
