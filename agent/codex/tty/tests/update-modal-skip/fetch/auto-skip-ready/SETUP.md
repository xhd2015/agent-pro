# Scenario

**Bug**: usage times out on Update available; must auto-Skip and fetch

```
fake-tui-auto-skip.py -> CSI Down + Enter (production) -> usage fields ready
```

## Preconditions

1. Fake starts on Update now, moves to Skip on CSI Down, idle on Enter.
2. Daemon-like PATH + isolated `TTY_WATCH_HOME`.

## Steps

1. `ShowStatusCommand` = auto-skip fake.
2. Unique `SessionID`.
3. `FetchTimeoutSecs=30`.

## Context

Happy path for auto-Skip protocol.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ShowStatusCommand = autoSkipFakeCommand(d, req)
	req.SessionID = "codex-update-modal-auto-skip"
	req.MarkerDir = filepath.Join(t.TempDir(), "markers")
	req.FetchTimeoutSecs = 30
	return nil
}
```
