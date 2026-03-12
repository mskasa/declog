# 0003: MADR Format Compatibility

- Date: 2026-03-12
- Status: Accepted
- Author: masahiro.kasatani

## Context

ADR (Architectural Decision Record) tools use various Markdown formats.
Choosing an idiosyncratic format would make it harder to migrate to or integrate with other tools.
Teams already using ADR tooling should be able to adopt declog with minimal friction.

## Decision

Adopt a template compatible with [MADR (Markdown Architectural Decision Records)](https://adr.github.io/madr/).
The template includes sections for Context, Decision, Consequences, and Alternatives Considered.

## Consequences

- Decisions written with declog are readable by other MADR-compatible tools
- The format is human-readable and requires no special tooling to view
- Consistent structure makes it easier to scan and compare decisions across a project
- The template is intentionally minimal — teams can extend sections as needed

## Alternatives Considered

- **Nygard format (original ADR format):** Simpler (Context, Decision, Status, Consequences) but less expressive; no Alternatives section
- **Custom format:** Maximum flexibility but no interoperability with existing tooling
- **YAML/TOML front matter:** Machine-readable metadata but less readable inline; adds parsing complexity
