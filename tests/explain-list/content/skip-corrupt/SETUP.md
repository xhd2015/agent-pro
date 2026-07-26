# Scenario

**Feature**: corrupt session directories are skipped silently

```
# corrupt dir (no session.data) + valid dir
explain list -> only valid session listed; exit 0; no crash noise required
```

## Preconditions

- One corrupt dir (harness creates dirname only, no session.data).
- One valid session with known Q text.

## Steps

1. Seed corrupt + valid sessions.
2. Run list.
3. Assert only valid content; title 1 shown; no failure.

## Context

- Matches existing `listSessions` skip-on-read-error behavior.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"list"}
	req.Sessions = []SessionSeed{
		// Corrupt: dir only, no session.data (convention: empty fields + nil Messages).
		{DirName: "2026-07-13-08-00-00-corrupt-eeeeeeee"},
		// Also a dir with bad JSON would be skipped — covered by missing file case.
		simpleSession(
			"2026-07-13-09-00-00-valid-ffffffff",
			"opencode", "deepseek-chat",
			"valid question", "valid answer",
		),
	}
	return nil
}
```
