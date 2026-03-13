Output docs/demo.gif

Set FontSize 14
Set Width 1000
Set Height 700
Set Theme "TokyoNight"
Set Padding 20
Set Framerate 24
Set PlaybackSpeed 0.85

# --- Hidden setup ---
Hide

# Build the why binary from this repo
Type "go build -o /tmp/why . 2>/dev/null"
Enter
Sleep 6s

# Set PATH and disable editor (use true as no-op)
Type "export PATH=/tmp:$PATH && export EDITOR=true && export PS1='$ '"
Enter
Sleep 300ms

# Create demo directory
Type "rm -rf /tmp/demo-app && mkdir -p /tmp/demo-app/docs/decisions /tmp/demo-app/internal /tmp/demo-app/database"
Enter
Sleep 300ms

Type "cd /tmp/demo-app && git init -q && git config user.email 'demo@example.com' && git config user.name 'Alice'"
Enter
Sleep 500ms

# ADR 0001: Switch to Redis (mentions internal/session.go for why blame demo)
Type "printf '%s' $'# 0001: Switch to Redis for session storage\n\n- Date: 2025-12-01\n- Status: Active\n- Author: Alice\n\n## Context\n\nSession data was stored in-process, causing issues across instances.\n\n## Decision\n\nUse Redis as a shared session store.\n\n## Related Files\n\ninternal/session.go\n' > docs/decisions/0001-switch-to-redis-for-session-storage.md"
Enter
Sleep 300ms

# ADR 0002: Connection pool (mentions database/connection.go)
Type "printf '%s' $'# 0002: Set database connection pool size to 10\n\n- Date: 2025-12-15\n- Status: Active\n- Author: Bob\n\n## Context\n\nConnection exhaustion caused timeouts under load.\n\n## Decision\n\nSet pool size to 10 based on load testing results.\n\n## Related Files\n\ndatabase/connection.go\n' > docs/decisions/0002-set-database-connection-pool-size-to-10.md"
Enter
Sleep 300ms

# ADR 0003: Cobra
Type "printf '%s' $'# 0003: Use cobra for CLI framework\n\n- Date: 2026-01-10\n- Status: Active\n- Author: Alice\n\n## Context\n\nNeeded a CLI framework with subcommand support and shell completion.\n\n## Decision\n\nUse cobra, the de facto standard for Go CLIs.\n' > docs/decisions/0003-use-cobra-for-cli-framework.md"
Enter
Sleep 300ms

# Demo source files (referenced in why blame)
Type "touch internal/session.go database/connection.go"
Enter
Sleep 300ms

Type "clear"
Enter
Sleep 200ms

Show

# === Demo ===

# 1. why log — show file creation and generated template
Type `why log "Add rate limiting to API endpoints"`
Enter
Sleep 1200ms

Type "cat docs/decisions/0004-add-rate-limiting-to-api-endpoints.md"
Enter
Sleep 4000ms

# 2. why list — show all decisions
Type "why list"
Enter
Sleep 3500ms

# 3. why search — keyword search
Type "why search Redis"
Enter
Sleep 3500ms

# 4. why blame — reverse lookup by file path
Type "why blame internal/session.go"
Enter
Sleep 3500ms
