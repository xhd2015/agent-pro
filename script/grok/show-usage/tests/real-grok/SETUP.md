# Scenario

**Profile**: `real-grok` — production `grok` interactive TUI on PATH (`label: real-grok, slow`)

```
# no GROK_SHOW_USAGE_COMMAND -> real grok in PTY -> /usage show -> live usage lines
grok-show-usage -> real grok -> usage patterns (not exact reset date)
```

## Preconditions

- Real `grok` CLI must be on `PATH`; otherwise tests skip with `grok not found in PATH`.
- Do **not** set `GROK_SHOW_USAGE_COMMAND` — production grok argv only.
- Leaves require `doctest test --label real-grok` (excluded from default runs).
- Longer timeouts acceptable for real TUI startup.

## Steps

1. Grouping `Setup` calls `exec.LookPath("grok")` and sets `req.SkipFakeCommand = true`.
2. Leaf runs against live grok and asserts stdout line patterns.
3. Assert stderr is empty on success.

## Context

- Real-grok leaf asserts structure (`Weekly limit: <digits>%`, `Next reset: <non-empty>`),
  not exact reset date (live grok changes daily).

```go
import (
	"os/exec"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if _, err := exec.LookPath("grok"); err != nil {
		t.Skip("grok not found in PATH")
	}
	req.SkipFakeCommand = true
	req.TimeoutSeconds = "60"
	return nil
}
```