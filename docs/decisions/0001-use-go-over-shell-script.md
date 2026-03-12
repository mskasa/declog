# 0001: Use Go Over Shell Script

- Date: 2026-03-12
- Status: Accepted
- Author: masahiro.kasatani

## Context

declog needs to run on Linux, macOS, and Windows as a single distributable tool.
Shell scripts are not portable across platforms, and Bash is not available on Windows without additional setup.
The tool also needs to handle file I/O, Git integration, and text processing reliably.

## Decision

Implement declog in Go instead of shell script.
Go compiles to a single static binary, supports all target platforms, and provides type safety and a rich standard library.

## Consequences

- Single binary distribution via GoReleaser — no runtime dependency required
- Windows support without WSL or Bash
- Type safety reduces a class of runtime bugs common in shell scripts
- Slightly higher barrier to contribution compared to shell scripts, but standard for CLI tooling in the Go ecosystem

## Alternatives Considered

- **Shell script (Bash):** Simple to write but not portable to Windows and harder to test reliably
- **Python:** Cross-platform but requires a Python runtime on the user's machine; packaging is more complex
- **Node.js:** Cross-platform but requires Node.js runtime; binary distribution requires bundling (e.g., pkg)
