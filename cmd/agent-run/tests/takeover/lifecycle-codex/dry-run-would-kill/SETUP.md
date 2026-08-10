# Scenario

**Feature**: `--dry-run` with live native codex plans kill + iTerm without side effects

```
seed Codex provider UUID
inject native codex pid 9202 hard-hit
  -> agent-run takeover --codex --dry-run UUID
  -> exit 0
  -> stdout: dry-run: would kill pid 9202 …
  -> stdout: dry-run: would open iTerm2 with: …
  -> kill log empty; no iTerm script file; no new meta
```

## Preconditions

- Same live-native fixture shape as kill-and-iterm, with `--dry-run`.
- Unmapped provider session (import would be the live path).

## Steps

1. Seed Codex + hooks pid 9202.
2. iTerm capture enabled (must remain empty for dry-run).
3. Args = `takeover --codex --dry-run <uuid>`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = "handle"
	providerID := takeoverCodexFixtureSessionID
	ws := absPath(t, req.TempDir)
	codexHome := takeoverCodexHome(req)
	seedCodexSession(t, codexHome, ws, providerID)

	const nativePID = 9202
	openPath := codexOpenPath(codexHome, providerID)
	writeTakeoverHooks(t, req,
		[]takeoverHookProc{
			{PID: nativePID, PPID: 1, Cmd: "/usr/local/bin/codex"},
		},
		map[int][]string{
			nativePID: {openPath},
		},
	)
	applyIterm2TestHooks(req)
	req.Args = []string{
		"takeover",
		"--codex",
		"--dry-run",
		providerID,
	}
	return nil
}
```
