# Scenario

**Feature**: live native grok PID (not under agent-run) → kill then iTerm ForceNew

```
seed Grok provider UUID (unmapped)
inject hooks:
  grok pid 9101 (ppid 1) + open-file hard hit on UUID
  -> agent-run takeover --grok UUID
  -> exit 0
  -> kill log records pid 9101 (SIGTERM→SIGKILL path)
  -> stdout mentions killed pid 9101
  -> iTerm ForceNew with agent-run follow-up
```

## Preconditions

- Native grok is **not** parented by agent-run / agent-run serve.
- Unmapped so follow-up is import-or-open after kill (either is OK if session-id printed).
- Kill goes through injectable Kill → kill_log (synthetic PID; no host process).

## Steps

1. Seed Grok session.
2. Hooks: single native grok hard-hit pid 9101.
3. iTerm capture enabled.
4. Args = `takeover --grok <uuid>`.

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

	const nativePID = 9101
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
		providerID,
	}
	return nil
}
```
