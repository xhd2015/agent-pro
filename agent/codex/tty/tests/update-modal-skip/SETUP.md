# Scenario

**Feature**: FetchStatus auto-Skip Codex update menu; ParseStatusSnapshot fields

```
# parse path
signed status fixture -> ParseStatusSnapshot -> monthly/credits/reset

# fetch path
CODEX_SHOW_STATUS_COMMAND fake TUI
  -> FetchStatus waitForPrompt auto-Skip (CSI Down + verify + Enter)
  -> /status -> UsageInfo
```

## Preconditions

1. Nested root under `agent/codex/tty/tests/update-modal-skip`.
2. Fixtures + fake scripts copied into `testdata/update-modal-skip/`.
3. Production Skip protocol in `waitForPrompt` / `dismissBlockingUpdateMenu`.
4. `python3` available for fetch leaves.

## Steps

1. Root `Setup` sets default timeout and fixtures dir.
2. Leaf sets `Op` (parse|fetch) and fixture or fake command.
3. Leaf `Assert` checks fields or FetchOK/error + markers.

## Context

Source-module ownership for Skip protocol. Classifier leaves live under
`pkgs/agenttty/tests/update-menu/`.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.FixturesDir == "" {
		req.FixturesDir = filepath.Join(d.DOCTEST_ROOT, "testdata", "update-modal-skip")
	}
	if req.FetchTimeoutSecs <= 0 {
		req.FetchTimeoutSecs = 30
	}
	return nil
}
```
