package context

import (
	"strings"
	"testing"
)

func TestSyncBlock_AppendsWhenMarkersAbsent(t *testing.T) {
	content := "# My CLAUDE.md\n\nSome hand-written notes.\n"
	block := manifestStartMarker + "\ngenerated\n" + manifestEndMarker

	got := SyncBlock(content, block)
	if !strings.HasPrefix(got, content) {
		t.Errorf("expected original content preserved as prefix, got: %q", got)
	}
	if !strings.Contains(got, block) {
		t.Errorf("expected block appended, got: %q", got)
	}
}

func TestSyncBlock_ReplacesExistingBlockInPlace(t *testing.T) {
	content := "# Notes\n\nBefore.\n\n" + manifestStartMarker + "\nold\n" + manifestEndMarker + "\n\nAfter.\n"
	newBlock := manifestStartMarker + "\nnew\n" + manifestEndMarker

	got := SyncBlock(content, newBlock)
	if !strings.Contains(got, "Before.") || !strings.Contains(got, "After.") {
		t.Errorf("expected surrounding hand-written content preserved, got: %q", got)
	}
	if strings.Contains(got, "old") {
		t.Errorf("expected old block content removed, got: %q", got)
	}
	if !strings.Contains(got, "new") {
		t.Errorf("expected new block content present, got: %q", got)
	}
}

func TestSyncBlock_Idempotent(t *testing.T) {
	content := "# Notes\n"
	block := manifestStartMarker + "\ngenerated\n" + manifestEndMarker

	once := SyncBlock(content, block)
	twice := SyncBlock(once, block)
	if once != twice {
		t.Errorf("expected idempotent result, got once=%q twice=%q", once, twice)
	}
}

// TestSyncBlock_IgnoresMarkerTextInsideGeneratedContent is a regression test: this repo's own
// agent-manifest-sync-format.md ADR quotes "<!-- kizami:start -->"/"<!-- kizami:end -->" as
// example text in its Decision section, which ends up embedded mid-line inside a manifest table
// row. A plain substring search for the end marker would match that decoy instead of the real
// terminator, truncating the captured block and making sync/--check disagree with themselves
// immediately after a sync.
func TestSyncBlock_IgnoresMarkerTextInsideGeneratedContent(t *testing.T) {
	block := manifestStartMarker + "\n" +
		"| ... | example: `" + manifestStartMarker + " ... " + manifestEndMarker + "` | slug |\n" +
		manifestEndMarker

	content := "# Notes\n"
	synced := SyncBlock(content, block)
	if !HasCurrentBlock(synced, block) {
		t.Fatalf("expected HasCurrentBlock to recognize its own just-written block, got content: %q", synced)
	}

	// A second sync with the identical block must be a no-op, not truncate at the decoy marker.
	again := SyncBlock(synced, block)
	if again != synced {
		t.Errorf("expected idempotent result despite in-content decoy markers, got: %q", again)
	}
}

func TestHasCurrentBlock(t *testing.T) {
	block := manifestStartMarker + "\ngenerated\n" + manifestEndMarker

	if HasCurrentBlock("# Notes\n", block) {
		t.Error("expected false when markers are absent")
	}

	synced := SyncBlock("# Notes\n", block)
	if !HasCurrentBlock(synced, block) {
		t.Error("expected true right after sync")
	}

	staleBlock := manifestStartMarker + "\nchanged\n" + manifestEndMarker
	if HasCurrentBlock(synced, staleBlock) {
		t.Error("expected false when block content differs")
	}
}
