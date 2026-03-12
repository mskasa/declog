# 0006: Command Name "why"

- Date: 2026-03-12
- Status: Accepted
- Author: masahiro.kasatani

## Context

The CLI tool needs a command name that is short, memorable, and communicates its purpose.
The repository is named `declog` (decision log), but the invocation name is a separate concern.
The command is used frequently during development, so brevity matters.

## Decision

Name the CLI command `why` instead of `dec`, `declog`, or `adr`.
"why" directly expresses the intent of the tool: recording the reasoning behind decisions.

## Consequences

- The command reads naturally in context: `why log "use postgres"`, `why list`, `why show 3`
- Short enough to type frequently without friction
- Memorable and self-documenting — new team members immediately understand what the tool is for
- `why` is an uncommon command name, so collisions with existing system commands are unlikely
- The binary is distributed as `why`, while the repository remains `declog` to avoid ambiguity in search and package management

## Alternatives Considered

- **`declog`:** Matches the repository name but is longer and less expressive as a verb
- **`dec`:** Short but ambiguous (decimal? declare?)
- **`adr`:** Describes the artifact type but not the action; also conflicts with the existing `adr-tools` project
- **`record`:** Descriptive but too generic and likely to conflict with other tools

## Related Files

- `main.go`
- `.goreleaser.yaml`
