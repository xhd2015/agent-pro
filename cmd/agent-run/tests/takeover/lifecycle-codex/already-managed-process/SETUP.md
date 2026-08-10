# Scenario

**Feature**: Codex provider already managed via process ancestry under agent-run

```
seed Codex provider UUID
inject ListProcs/Lsof:
  agent-run serve (ppid root) -> child codex with open-file hard hit on rollout UUID
  -> agent-run takeover --codex UUID
  -> exit 0
  -> warning: already managed
  -> no kill log; no iTerm
```

## Preconditions

- Provider session exists under CODEX_HOME.
- Hooks snapshot models agent-run parent + codex child hard-hit (no real PIDs).
- Open path is rollout under `…/.codex/sessions/…/rollout-…-<uuid>.jsonl`.
- Optional: no agent-run meta mapping (process ancestry alone is enough).

## Steps

1. Seed Codex session.
2. Write hooks: agent-run serve pid + codex child with rollout open path.
3. Enable iTerm capture (must stay empty).
4. Args = `takeover --codex <uuid>`.

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

	const (
		agentRunPID = 9800
		codexPID    = 9801
	)
	openPath := codexOpenPath(codexHome, providerID)
	writeTakeoverHooks(t, req,
		[]takeoverHookProc{
			{PID: agentRunPID, PPID: 1, Cmd: "/usr/local/bin/agent-run serve --session " + providerID},
			{PID: codexPID, PPID: agentRunPID, Cmd: "/usr/local/bin/codex"},
		},
		map[int][]string{
			codexPID: {openPath},
		},
	)
	applyIterm2TestHooks(req)
	req.Args = []string{
		"takeover",
		"--codex",
		providerID,
	}
	return nil
}
```
