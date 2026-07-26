# Scenario

**Feature**: successful usage fetch prints exactly two stdout lines

```
# fake TUI returns usage after /usage show -> exit 0, stdout has Weekly limit + Next reset
grok-show-usage (GROK_SHOW_USAGE_COMMAND=fake) -> PTY -> /usage show -> parse -> print
doctest <- stdout: two lines only; stderr empty
```

## Preconditions

- `GROK_SHOW_USAGE_COMMAND` is set to a fake TUI that responds to `/usage show`.
- Fake TUI prints `Grok ›` before reading stdin.

## Steps

1. Grouping `Setup` sets default fake TUI hook unless leaf overrides.
2. Leaf `Setup` may customize fixture values.
3. Run CLI and assert exit 0, empty stderr, exact stdout lines.

## Context

- Success leaves use `assert.Output` for exact stdout matching.
- Default fake fixture: `1%` and `July 9, 16:55 PT`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Bin == "" {
		t.Fatalf("success setup: grok-show-usage binary not built (root Setup skipped?)")
	}
	req.SkipFakeCommand = false
	if req.ShowUsageCommand == "" {
		req.ShowUsageCommand = fakeTUIDefault()
	}
	return nil
}
```