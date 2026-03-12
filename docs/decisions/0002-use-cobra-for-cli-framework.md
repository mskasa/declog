# 0002: Use Cobra for CLI Framework

- Date: 2026-03-12
- Status: Accepted
- Author: masahiro.kasatani

## Context

declog exposes multiple subcommands (`log`, `list`, `search`, `show`, `status`).
Implementing subcommand routing, flag parsing, and help text generation from scratch would be repetitive and error-prone.
A CLI framework is needed to handle these concerns consistently.

## Decision

Use [cobra](https://github.com/spf13/cobra) as the CLI framework.
Cobra is the de facto standard for Go CLI applications and provides subcommand management, automatic help generation, and shell completion out of the box.

## Consequences

- Subcommand structure and flag parsing are handled consistently without boilerplate
- Shell completion (Bash, Zsh, Fish, PowerShell) is available with minimal effort
- Widely used in the Go ecosystem (Kubernetes, Hugo, GitHub CLI), so contributors are likely familiar with it
- Adds an external dependency, but it is mature and well-maintained

## Alternatives Considered

- **`flag` (stdlib):** No subcommand support; too low-level for a multi-command CLI
- **`urfave/cli`:** A valid alternative but less widely adopted than cobra in the Go ecosystem
- **Manual routing:** Simple `switch` on `os.Args` — sufficient for a few commands but does not scale and lacks help/completion generation

## Related Files

- `cmd/root.go`
- `cmd/log.go`
- `cmd/list.go`
- `cmd/search.go`
- `cmd/show.go`
- `cmd/status.go`
