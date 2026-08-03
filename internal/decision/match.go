package decision

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

// MatchKind identifies which Related Files matching rule matched a candidate file.
// Decision to unify Blame/CheckHook/Audit behind this single definition:
// docs/decisions/2026-08-03-related-files-single-definition.md
type MatchKind string

const (
	MatchExact MatchKind = "exact"
	MatchDir   MatchKind = "dir"
	MatchGlob  MatchKind = "glob"
)

// Match reports whether entry (a Related Files list item) matches file (a repo-relative path),
// and if so, which kind of rule matched.
//
// entry is tried as a glob first (if it contains "*" or "?"). Otherwise, a trailing "/" is
// stripped (cosmetic only — it does not change matching behavior) and entry matches file either
// exactly, or as a directory prefix when file has entry as a path-component prefix. The prefix
// form applies even when entry has no trailing slash: docs/decisions/2026-08-03-related-files-single-definition.md
func Match(entry, file string) (MatchKind, bool) {
	entry = filepath.ToSlash(entry)
	file = filepath.ToSlash(file)

	if strings.ContainsAny(entry, "*?") {
		if globMatch(entry, file) {
			return MatchGlob, true
		}
		return "", false
	}

	dir := strings.TrimSuffix(entry, "/")
	switch {
	case file == dir:
		return MatchExact, true
	case strings.HasPrefix(file, dir+"/"):
		return MatchDir, true
	default:
		return "", false
	}
}

// globMatch reports whether file matches pattern, where "**" matches zero or more whole
// path segments and "*"/"?" match within a single segment (path.Match semantics).
// Go's stdlib path.Match has no "**", so this is implemented as a small recursive matcher
// rather than pulling in a dependency: docs/decisions/2026-08-03-related-files-single-definition.md
func globMatch(pattern, file string) bool {
	return globMatchSegments(strings.Split(pattern, "/"), strings.Split(file, "/"))
}

func globMatchSegments(pattern, file []string) bool {
	if len(pattern) == 0 {
		return len(file) == 0
	}
	if pattern[0] == "**" {
		if globMatchSegments(pattern[1:], file) {
			return true
		}
		if len(file) == 0 {
			return false
		}
		return globMatchSegments(pattern, file[1:])
	}
	if len(file) == 0 {
		return false
	}
	ok, err := path.Match(pattern[0], file[0])
	if err != nil || !ok {
		return false
	}
	return globMatchSegments(pattern[1:], file[1:])
}

// EntryExists reports whether entry (a Related Files list item) still points at something
// real under repoRoot. Glob entries are not verified (always reported as existing) — checking
// whether a glob still matches at least one real file requires a directory walk, which is out
// of scope for now: docs/decisions/2026-08-03-related-files-single-definition.md
func EntryExists(repoRoot, entry string) bool {
	if strings.ContainsAny(entry, "*?") {
		return true
	}
	p := strings.TrimSuffix(entry, "/")
	_, err := os.Stat(filepath.Join(repoRoot, p))
	return err == nil
}
