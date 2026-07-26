# Scenario

**Feature**: `--open` on real Grok under `llm-mock-run-grok` (Option C harness)

```
# shared harness (no hooks)
build llm-mock-run-grok + agent-run
  -> isolated AGENT_RUN_HOME + GROK_HOME + workspace
  -> strip AGENT_RUN_GROK_TTY_COMMAND + LLM_MOCK_RUN_GROK_COMMAND
  -> LLM_MOCK_RUN_FLAGS=--log-http <tmp>.jsonl
  -> AGENT_RUN_OPEN_ATTACH_INSTANT=1
  -> requires real grok on PATH (label: real-grok)

# leaves split submit mode
--open --no-submit "draft" -> no model turn (no-model-turn)
--open "draft"             -> model turn evidence (control-model-turn)
```

## Preconditions

- Real `grok` CLI on `PATH`; otherwise leaves skip (`t.Skip`).
- Leaves use run-profile labels `real-grok` and `slow` so default CI can omit them.
- `--agent-runner-binary` points at session-built `llm-mock-run-grok`.
- `--agent-runner-config-home` is a per-leaf isolated directory (also used as
  GROK_HOME for the mock child via `env GROK_HOME=…`).
- No fake TUI: hooks must remain unset for the Option C oracle.

## Steps

1. Grouping `Setup` calls `configureRealGrokOpen` (LookPath grok, strip hooks,
   log-http, instant attach, cleanup registry PIDs).
2. Leaves set `NoSubmit` true/false and finalize `req.Args`.
3. `Run` with `Mode=open-real-grok-after` settles then fills turn oracles.
4. Assert checks no-turn vs control-turn evidence.

## Context

- Default draft: `draft-no-submit-OPTIONC-probe-zz9` (root constant).
- Settle default: 5s after CLI exit so a buggy argv auto-submit would have
  produced HTTP/session evidence.
- Exec timeout default: 120s (open bind post-detach grace can be ~20s).

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if err := configureRealGrokOpen(t, req); err != nil {
		return err
	}
	// Default to no-submit core path; control leaf clears NoSubmit.
	req.NoSubmit = true
	req.Args = openRealGrokArgs(req)
	req.ExecTimeout = defaultRealGrokTimeout
	if req.SettleAfter <= 0 {
		req.SettleAfter = defaultSettleAfter
	}
	return nil
}
```
