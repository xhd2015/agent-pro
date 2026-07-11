# Scenario

**Feature**: if Skip cannot be selected, fail without Entering Update now

```
fake-tui-stuck-update-now.py (CSI Down ignored) -> error; never ready
```

## Preconditions

Fake always keeps `›` on **1. Update now**. CSI Down is a no-op.

## Steps

1. `ShowStatusCommand` = stuck fake.
2. `MarkerDir` records illegal Enter.
3. `FetchTimeoutSecs=15` (may still wait until early error/timeout).

## Context

Negative contract: never silent upgrade.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.ShowStatusCommand = stuckUpdateNowFakeCommand(req)
	req.SessionID = "codex-update-modal-stuck"
	req.MarkerDir = filepath.Join(t.TempDir(), "markers")
	req.FetchTimeoutSecs = 15
	return nil
}
```
