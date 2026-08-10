# Scenario

**Feature**: provider session already managed via process ancestry under agent-run

```
seed Grok provider UUID
inject ListProcs/Lsof:
  agent-run serve (ppid root) -> child grok with open-file hard hit on UUID
  -> agent-run takeover --grok UUID
  -> exit 0
  -> warning: already managed
  -> no kill log; no iTerm
```

## Preconditions

- Provider session exists under GROK_HOME (so failure is managed-gate, not missing).
- Hooks snapshot models agent-run parent + grok child hard-hit (no real PIDs).
- Optional: no agent-run meta mapping (process ancestry alone is enough).

## Steps

1. Seed Grok session.
2. Write hooks: agent-run serve pid + grok child with session open path.
3. Enable iTerm capture (must stay empty).
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

	const (
		agentRunPID = 8800
		grokPID     = 8801
	)
	openPath := grokOpenPath(grokHome, ws, providerID)
	writeTakeoverHooks(t, req,
		[]takeoverHookProc{
			{PID: agentRunPID, PPID: 1, Cmd: "/usr/local/bin/agent-run serve --session " + providerID},
			{PID: grokPID, PPID: agentRunPID, Cmd: "/usr/local/bin/grok"},
		},
		map[int][]string{
			grokPID: {openPath},
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
