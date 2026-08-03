package context

import "strings"

// SyncBlock returns content with its kizami-owned marker-delimited section replaced by
// block, or block appended (with a blank-line separator) if content has no marker pair yet.
// Only the marked span is ever touched — hand-written content elsewhere is left alone:
// docs/decisions/2026-08-03-agent-manifest-sync-format.md
func SyncBlock(content, block string) string {
	start := indexOfStandaloneLine(content, manifestStartMarker, 0)
	if start == -1 {
		return appendBlock(content, block)
	}
	end := indexOfStandaloneLine(content, manifestEndMarker, start+len(manifestStartMarker))
	if end == -1 {
		return appendBlock(content, block)
	}
	end += len(manifestEndMarker)
	return content[:start] + block + content[end:]
}

// HasCurrentBlock reports whether content's existing marker-delimited section already
// equals block exactly (used by `kizami agents sync --check`).
func HasCurrentBlock(content, block string) bool {
	start := indexOfStandaloneLine(content, manifestStartMarker, 0)
	if start == -1 {
		return false
	}
	end := indexOfStandaloneLine(content, manifestEndMarker, start+len(manifestStartMarker))
	if end == -1 {
		return false
	}
	end += len(manifestEndMarker)
	return content[start:end] == block
}

func appendBlock(content, block string) string {
	trimmed := strings.TrimRight(content, "\n")
	if trimmed == "" {
		return block + "\n"
	}
	return trimmed + "\n\n" + block + "\n"
}

// indexOfStandaloneLine returns the byte offset (at or after from) of marker, but only when
// marker occupies an entire line by itself (nothing else before or after it on that line).
// A plain substring search would also match marker text that merely appears inside generated
// content — e.g. a decision's summary that quotes "<!-- kizami:start -->" as an example, which
// is exactly what this repo's own agent-manifest-sync-format.md ADR does. Returns -1 if no
// standalone occurrence is found.
func indexOfStandaloneLine(content, marker string, from int) int {
	if from > len(content) {
		return -1
	}
	search := content[from:]
	offset := 0
	for {
		idx := strings.Index(search[offset:], marker)
		if idx == -1 {
			return -1
		}
		pos := from + offset + idx
		lineStart := pos == 0 || content[pos-1] == '\n'
		after := pos + len(marker)
		lineEnd := after == len(content) || content[after] == '\n'
		if lineStart && lineEnd {
			return pos
		}
		offset += idx + 1
	}
}
