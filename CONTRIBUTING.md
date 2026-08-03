# Contributing to kizami

Thank you for your interest in contributing!
This document explains how to get started and what to keep in mind when submitting changes.

---

## Table of Contents

- [Local Development Setup](#local-development-setup)
- [Running Tests](#running-tests)
- [Running the Linter](#running-the-linter)
- [Commit Message Convention](#commit-message-convention)
- [Submitting a Pull Request](#submitting-a-pull-request)
- [Recording Design Decisions](#recording-design-decisions)
- [Code of Conduct](#code-of-conduct)

---

## Local Development Setup

This project uses [mise](https://mise.jdx.dev/) to manage tool versions.

### 1. Install mise

```bash
curl https://mise.run | sh
```

See the [official docs](https://mise.jdx.dev/getting-started.html) for other installation methods (Homebrew, etc.).

### 2. Install tools

```bash
mise install
```

This installs the exact versions of Go and golangci-lint defined in `.mise.toml`.

### 3. Verify

```bash
go version           # go1.25
golangci-lint version  # v2.12.2
```

---

## Running Tests

```bash
go test ./...
```

Tests that depend on [ripgrep](https://github.com/BurntSushi/ripgrep) are skipped automatically when `rg` is not installed.

---

## Running the Linter

```bash
golangci-lint run
```

The linter configuration is in `.golangci.yml`.

---

## Commit Message Convention

```
<type>: <summary>

Types:
  feat     New feature
  fix      Bug fix
  docs     Documentation changes (including ADRs)
  refactor Refactoring without behaviour change
  test     Adding or updating tests
  chore    Build, dependency, or tooling changes
```

Examples:

```
feat: add kizami archive command
fix: handle slug collision across multiple directories
docs: add ADR for archive command design
```

---

## Submitting a Pull Request

1. Fork the repository and create a branch from `main`.
2. Make your changes, including tests where appropriate.
3. Ensure `go test ./...` and `golangci-lint run` both pass.
4. Open a pull request against `main` using the provided PR template.

For significant design changes, please open an issue first to discuss the approach.

---

## Recording Design Decisions

kizami uses itself to record its own design decisions (dogfooding).
If your contribution involves a meaningful design choice — a new command, a change in behaviour, a trade-off between approaches — please add an ADR under `docs/decisions/`.

See the existing ADRs for examples of the expected format.

---

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md).
Please read it before participating.
