# Scenario

**Feature**: provider session already managed via live agent-run registry → soft no-op

```
seed Grok provider UUID under GROK_HOME
seed agent-run meta: runner=grok-tty, runner_session_id=UUID, status=running
seed live grok-tty-registry entry for agent session id (real sleep PID)
  -> agent-run takeover --grok UUID
  -> exit 0
  -> stderr warning: already managed by agent-run
  -> sleep PID still alive; no iTerm script; no kill log
```

## Preconditions

- Mapping: `meta.runner_session_id` = provider UUID, `meta.runner` = `grok-tty`.
- Registry live for the **agent-run** session id (not the provider UUID).
- Empty injectable hooks (registry path does not need ListProcs).

## Steps

1. Seed Grok session + mapped running meta.
2. Start sleep + write registry for agent session id.
3. Enable iTerm capture (must stay empty) + empty hooks.
4. Args = `takeover --grok <provider-uuid>`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = "handle"
	providerID := takeoverFixtureSessionID
	agentID := "takeover-managed-reg-s1"
	ws := absPath(t, req.TempDir)
	seedGrokSession(t, takeoverGrokHome(req), ws, providerID)
	seedMappedMeta(t, req, "grok-tty", agentID, providerID, "running")
	pid := startLiveSleepFixture(t, req)
	writeTakeoverRegistryEntry(t, req.Home, agentID, pid)
	writeEmptyTakeoverHooks(t, req)
	applyIterm2TestHooks(req)
	req.Args = []string{
		"takeover",
		"--grok",
		providerID,
	}
	return nil
}
```
