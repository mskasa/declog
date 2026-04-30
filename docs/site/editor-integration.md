---
layout: default
title: Editor Integration
nav_order: 6
---

# Editor Integration

kizami provides a VS Code extension that brings related documents directly into your editor — no need to switch to the terminal to run `kizami blame`.

[← Back to Documentation](.)

---

## VS Code Extension

The **kizami** extension for VS Code shows which ADRs and design documents mention the file you are currently editing.

### Installation

Search for **kizami** in the VS Code Extensions Marketplace, or open the Extensions panel (`Ctrl+Shift+X` / `Cmd+Shift+X`) and search for `kizami`.

**Requirements:**
- kizami CLI installed and available on `PATH` (or configured via `kizami.binaryPath`)
- A workspace containing `kizami.toml` (created by `kizami init`)

### Features

#### Related Documents sidebar

When you open a source file, the **kizami** panel in the Activity Bar automatically lists every document that references the current file in its `## Related Files` section or `.kizami` sidecar.

Clicking a document opens it in VS Code's built-in Markdown preview.

#### Open in Editor

Right-click a document item in the sidebar and select **Open in Editor** to open it as a plain text file for editing.

#### Explorer context menu

Right-click any file in the Explorer panel and select **Find Related kizami Documents** to show documents related to that file without opening it first.

#### Refresh

Use the refresh button in the sidebar title bar to manually re-run `kizami blame` if documents were updated outside of VS Code.

---

### Configuration

| Setting | Default | Description |
|---|---|---|
| `kizami.binaryPath` | `"kizami"` | Path to the kizami binary. Set this if kizami is not on your `PATH`. |

To change the setting, open VS Code Settings (`Ctrl+,` / `Cmd+,`) and search for `kizami`.

---

### How it works

The extension runs `kizami blame <file>` in the background whenever you switch to a new file. It parses the output and populates the sidebar with the results. No data leaves your machine — everything runs locally using your installed kizami binary.

---

### Links

- [kizami-vscode on GitHub](https://github.com/mskasa/kizami-vscode)
- [kizami CLI on GitHub](https://github.com/mskasa/kizami)
