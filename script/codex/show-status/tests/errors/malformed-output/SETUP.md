# Scenario

**Feature**: parse failure when TUI output lacks status fields

```
# fake TUI prints garbage after /status -> parse error
codex-show-status -> fake codex (no Monthly credit limit) -> parse failure
```

## Preconditions

- Fake TUI responds to `/status` but prints unparseable output.

## Steps

1. Set `ShowStatusCommand` to `fakeTUIMalformed()`.
2. Assert non-zero exit; stderr mentions `parse`.

## Context

- Distinguishes parse errors from timeout (TUI responds quickly with bad data).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ShowStatusCommand = fakeTUIMalformed()
	return nil
}
```