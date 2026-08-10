# Scenario

**Feature**: `--dry-run` with live native grok plans kill + iTerm without side effects

```
seed Grok provider UUID
inject native grok pid 9102 hard-hit
  -> agent-run takeover --grok --dry-run UUID
  -> exit 0
  -> stdout: dry-run: would kill pid 9102 …
  -> stdout: dry-run: would open iTerm2 with: …
  -> kill log empty; no iTerm script file; no new meta
```

## Preconditions

- Same live-native fixture shape as kill-and-iterm, with `--dry-run`.
- Unmapped provider session (import would be the live path).

## Steps

1. Seed Grok + hooks pid 9102.
2. iTerm capture enabled (must remain empty for dry-run).
3. Args = `takeover --grok --dry-run <uuid>`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = "handle"
	providerID := takeoverFixtureSessionID
	ws := absPath(t, req.TempDir)
	grokHome := takeoverGrokHome(req)
	seedGrokSession(t, grokHome, ws, providerID)

	const nativePID = 9102
	openPath := grokOpenPath(grokHome, ws, providerID)
	writeTakeoverHooks(t, req,
		[]takeoverHookProc{
			{PID: nativePID, PPID: 1, Cmd: "/usr/local/bin/grok"},
		},
		map[int][]string{
			nativePID: {openPath},
		},
	)
	applyIterm2TestHooks(req)
	req.Args = []string{
		"takeover",
		"--grok",
		"--dry-run",
		providerID,
	}
	return nil
}
```
