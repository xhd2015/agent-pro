# Scenario

**Feature**: live native codex PID (not under agent-run) → kill then iTerm ForceNew

```
seed Codex provider UUID (unmapped)
inject hooks:
  codex pid 9201 (ppid 1) + open-file hard hit on rollout UUID
  -> agent-run takeover --codex UUID
  -> exit 0
  -> kill log records pid 9201 (SIGTERM→SIGKILL path)
  -> stdout mentions killed pid 9201 (codex)
  -> iTerm ForceNew with agent-run follow-up
```

## Preconditions

- Native codex is **not** parented by agent-run / agent-run serve.
- Unmapped so follow-up is import-or-open after kill.
- Kill goes through injectable Kill → kill_log (synthetic PID; no host process).

## Steps

1. Seed Codex session.
2. Hooks: single native codex hard-hit pid 9201.
3. iTerm capture enabled.
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

	const nativePID = 9201
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
		providerID,
	}
	return nil
}
```
