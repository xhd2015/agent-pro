# Scenario

**Feature**: `--dir` that differs from Grok session `info.cwd` is a hard error
(no relocate for import)

```
Grok info.cwd = WS_A
--dir WS_B (exists, ≠ WS_A)
  -> agent-run run --dir WS_B --resume-from-grok-session UUID
  -> exit 1; dir / cwd mismatch
```

## Preconditions

- Both WS_A and WS_B exist under the temp dir.
- Grok session is seeded with `info.cwd = WS_A`.
- No agent-run mapping for the UUID.

## Steps

1. Create `ws-a` and `ws-b` under `TempDir`.
2. Seed Grok session with cwd = `ws-a`.
3. Run with `--dir ws-b` (absolute).

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	wsA := filepath.Join(req.TempDir, "ws-a")
	wsB := filepath.Join(req.TempDir, "ws-b")
	mustMkdir(t, wsA)
	mustMkdir(t, wsB)
	req.GrokCWD = absPath(t, wsA)
	req.DirFlag = absPath(t, wsB)
	// Process cwd can be TempDir; gate compares --dir to Grok info.cwd only.
	req.WorkDir = req.TempDir
	seedGrokSession(t, req.GrokHome, req.GrokCWD, req.GrokSessionID)
	req.Args = runArgs(req, req.GrokSessionID)
	return nil
}
```
