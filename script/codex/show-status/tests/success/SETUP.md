# Scenario

**Feature**: successful status fetch prints exactly three stdout lines

```
# fake TUI returns status after /status -> exit 0, stdout has Monthly usage + Credits used + Next reset
codex-show-status (CODEX_SHOW_STATUS_COMMAND=fake) -> tty-watch -> /status -> parse -> print
doctest <- stdout: three lines only; stderr empty
```

## Preconditions

- `CODEX_SHOW_STATUS_COMMAND` is set to a fake TUI that responds to `/status`.
- Fake TUI prints `Codex ›` before reading stdin.

## Steps

1. Grouping `Setup` sets default fake TUI hook unless leaf overrides.
2. Leaf `Setup` may customize fixture values or assert registry cleanup.
3. Run CLI and assert exit 0, empty stderr, exact stdout lines.

## Context

- Success leaves use `assert.Output` for exact stdout matching.
- Default fake fixture: `58%` usage (`42% left`), `6519/11250` credits, `08:00 on 1 Aug` reset.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.Bin == "" {
		t.Fatalf("success setup: codex-show-status binary not built (root Setup skipped?)")
	}
	req.SkipFakeCommand = false
	if req.ShowStatusCommand == "" {
		req.ShowStatusCommand = fakeTUIDefault()
	}
	return nil
}
```