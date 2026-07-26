# Scenario

**Feature**: usage.Fetch Codex provider with CODEX_SHOW_STATUS_COMMAND fake TUI

```
usage.Fetch(ctx, Codex) -> codex tty -> in-process ttywatch -> fake TUI -> Snapshot
```

## Steps

1. Set `req.Provider` to `usage.Codex`.
2. Set `req.ShowStatusCommand` to default fake codex TUI fixture.
3. Use isolated `req.TTYWatchHome` from root Setup.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/agent/usage"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Provider = usage.Codex
	req.ShowStatusCommand = fakeCodexTUIDefault()
	return nil
}
```