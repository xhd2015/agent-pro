# Scenario

**Feature**: FetchStatus auto-Skip via CODEX_SHOW_STATUS_COMMAND fake TUI

```
fake update-menu TUI -> waitForPrompt Skip protocol -> /status -> UsageInfo
```

## Preconditions

1. `PATH` stripped to daemon-like PATH so only explicit python fake runs.
2. Fake scripts under `testdata/update-modal-skip/`.
3. Isolated `TTY_WATCH_HOME`.

## Steps

1. Set `Op=fetch`.
2. Leaf sets ShowStatusCommand (auto-skip or stuck) and MarkerDir.

## Context

End-to-end contract for production Skip keys (CSI Down + Enter).

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Op = "fetch"
	req.StripDaemonPATH = true
	if req.TTYWatchHome == "" {
		req.TTYWatchHome = filepath.Join(t.TempDir(), ".tty-watch")
	}
	if req.FetchTimeoutSecs <= 0 {
		req.FetchTimeoutSecs = 30
	}
	return nil
}
```
