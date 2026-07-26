# Scenario

**Feature**: timeout when fake TUI never prints status fields

```
# fake TUI hangs after prompt -> wait exceeds CODEX_SHOW_STATUS_TIMEOUT -> timeout error
codex-show-status -> fake codex (no status fields) -> timeout
```

## Preconditions

- Fake TUI prints prompt and reads `/status` but never emits `Monthly credit limit:`.
- `CODEX_SHOW_STATUS_TIMEOUT` shortened to 3 seconds for fast test.

## Steps

1. Set `ShowStatusCommand` to `fakeTUINoStatus()`.
2. Set `TimeoutSeconds` to `"3"`.
3. Assert non-zero exit; stderr mentions `timeout`.

## Context

- Default CLI timeout is 60s; this leaf overrides to keep CI fast.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ShowStatusCommand = fakeTUINoStatus()
	req.TimeoutSeconds = "3"
	return nil
}
```