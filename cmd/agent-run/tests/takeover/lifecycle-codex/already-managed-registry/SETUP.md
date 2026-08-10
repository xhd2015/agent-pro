# Scenario

**Feature**: Codex provider already managed via live codex-tty registry → soft no-op

```
seed Codex provider UUID under CODEX_HOME
seed agent-run meta: runner=codex-tty, runner_session_id=UUID, status=running
seed live codex-tty-registry entry for agent session id (real sleep PID)
  -> agent-run takeover --codex UUID
  -> exit 0
  -> stderr warning: already managed by agent-run
  -> sleep PID still alive; no iTerm script; no kill log
```

## Preconditions

- Mapping: `meta.runner_session_id` = provider UUID, `meta.runner` = `codex-tty`.
- Registry live under **codex-tty-registry/** for the agent-run session id.
- Empty injectable hooks (registry path does not need ListProcs).

## Steps

1. Seed Codex session + mapped running meta.
2. Start sleep + write codex-tty registry for agent session id.
3. Enable iTerm capture (must stay empty) + empty hooks.
4. Args = `takeover --codex <provider-uuid>`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = "handle"
	providerID := takeoverCodexFixtureSessionID
	agentID := "takeover-codex-managed-reg-s1"
	ws := absPath(t, req.TempDir)
	seedCodexSession(t, takeoverCodexHome(req), ws, providerID)
	seedMappedMeta(t, req, "codex-tty", agentID, providerID, "running")
	pid := startLiveSleepFixture(t, req)
	writeTakeoverCodexRegistryEntry(t, req.Home, agentID, pid)
	writeEmptyTakeoverHooks(t, req)
	applyIterm2TestHooks(req)
	req.Args = []string{
		"takeover",
		"--codex",
		providerID,
	}
	return nil
}
```
