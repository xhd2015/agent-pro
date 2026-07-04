# Scenario

**Bug**: `tty-watch snapshot codex` smears stacked alternate-screen redraws instead of
showing the latest codex UI state.

```
# codex-like alt-screen TUI with multiple full redraws
harness -> run --session-id codex (codex-like sh) -> detach -> snapshot codex
```

Reproduces user report: `tty-watch run --session-id codex codex` then
`tty-watch snapshot codex` prints overlapping MCP boot lines and duplicate UI boxes.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "snapshot-codex-like-single-screen"
	return nil
}
```