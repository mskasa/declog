# Add Related Files Section to ADR Template

- Date: 2026-03-12
- Status: Accepted
- Author: masahiro.kasatani

## Context

The planned `kizami blame <file>` command needs to find ADRs related to a given file path.
Without a dedicated section, the only option would be unstructured full-text search across the entire document, which produces noisy results and makes it harder to distinguish intentional file associations from incidental mentions.
A structured section in the template gives authors a clear place to record which files a decision affects, and gives `kizami blame` a reliable target to search.

## Decision

Add a mandatory `## Related Files` section to the ADR Markdown template.
The section is required to be present in every ADR, but its content may be left empty when the relationship to specific files is unclear or not applicable.

## Consequences

- `kizami blame <file>` can search the Related Files section for reliable, low-noise results
- Authors have a consistent, explicit place to document file relationships
- Existing ADRs (0001–0006) have been backfilled with the section; content was added where the relationship was clear and left empty otherwise
- The template grows by two lines, but remains minimal

## Alternatives Considered

- **Full-text search without a dedicated section:** Simpler template, but `kizami blame` would match incidental mentions (e.g. a file name appearing in an example) alongside intentional associations
- **Separate metadata file (e.g. YAML sidecar):** Machine-readable but splits the decision record across two files, violating the plaintext-only principle (ADR-0004)
- **Optional section:** Reduces noise in simple ADRs but makes `kizami blame` unreliable when the section is absent

## Related Files

- `internal/template/template.go`
- `cmd/blame.go`
