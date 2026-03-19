# Contributing

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
go version        # go1.23.2
golangci-lint version  # v1.64.8
```

## Running Tests

```bash
go test ./...
```

## Running Linter

```bash
golangci-lint run
```
