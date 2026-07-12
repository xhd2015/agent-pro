# Scenario

**Bug**: O1 — non-empty open prompt without `GROK_HOME` env still hard-waits for
delayed bind (not soft 750ms fail-fast)

```
# isolate default grok home via HOME; do NOT set GROK_HOME / session-id hook
HOME=<temp>/fake-home  (sessions under HOME/.grok)
  + non-empty open prompt
  + empty sessions at t=0
  + schedule updates.jsonl materialization after ~2s under HOME/.grok
  + AGENT_RUN_OPEN_ATTACH_INSTANT=1
  -> agent-run run --open "hard require via prompt"
  -> attach returns before discovery is possible
  -> open MUST wait past soft 750ms (hard path from non-empty prompt alone)
  -> bind succeeds after delayed materialization
  -> wall ≥ delay; stderr grok session + meta.runner_session_id; exit 0
```

Regression for soft 750ms race when user has default `~/.grok` and a real open
prompt (see `doc/LOOP_2026-07-11_open-bind-runner-unbound.md`). Existing B2
sets `GROK_HOME` (also hard-requires); this leaf proves **prompt alone** is enough.

## Preconditions

- Harness `NoGrokHomeEnv` strips ambient `GROK_HOME` / `AGENT_RUNNER_CONFIG_HOME` /
  `AGENT_RUN_GROK_TTY_GROK_SESSION_ID` from the child environ.
- Delayed materialization still writes under `req.GrokHome` (= `$HOME/.grok`).

## Steps

1. Point `HOME` at temp fake-home; set `GrokHome=$HOME/.grok` for materialize only.
2. Set `NoGrokHomeEnv=true`; non-empty open prompt; delay ~2s; no session-id hook env.
3. Run open; assert wait ≥ delay and bind success.

```go
import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

const hardRequireNoGrokHomeUUID = "550e8400-e29b-41d4-a716-446655440811"

// delay longer than soft 750ms and longer than instant-attach return.
const hardRequireNoGrokHomeDelay = 2 * time.Second

func Setup(t *testing.T, req *Request) error {
	prompt := "hard require via prompt"
	req.OpenPrompt = prompt
	req.InitialPrompt = prompt

	fakeHome := filepath.Join(req.TempDir, "fake-home")
	if err := os.MkdirAll(fakeHome, 0755); err != nil {
		return err
	}
	// Materialize under the default GrokHome path ($HOME/.grok) without setting GROK_HOME.
	req.GrokHome = filepath.Join(fakeHome, ".grok")
	if err := os.MkdirAll(req.GrokHome, 0755); err != nil {
		return err
	}
	req.GrokSessionUUID = hardRequireNoGrokHomeUUID
	req.GrokMaterializeDelay = hardRequireNoGrokHomeDelay
	req.NoGrokHomeEnv = true
	setEnvKV(req, "HOME", fakeHome)

	req.GrokTTYCommand = fakeTUIHoldSeconds(8)
	req.OpenInstantAttach = true
	req.Args = []string{"run", "--agent-runner", "grok-tty", "--open", prompt}
	req.ExecTimeout = 90 * time.Second
	return nil
}
```
