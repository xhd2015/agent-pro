# Scenario

**Feature**: control leaf — same Option C harness **without** `--no-submit`
proves the turn oracle is live (mock HTTP and/or session user_message)

```
agent-run run --agent-runner grok-tty --open "draft-no-submit-OPTIONC-probe-zz9"
  --agent-runner-binary llm-mock-run-grok
  --agent-runner-config-home <isolated>
  -> exit 0
  -> after settle: mock HTTP chat and/or session user_message for prompt
```

## Preconditions

- Same real-grok / llm-mock-run-grok harness as `no-model-turn` (no hooks).
- **Omits** `--no-submit` so product default auto-submits the open prompt
  (positional argv and/or inject with Enter).
- If this leaf fails to observe turn evidence, the no-model-turn leaf's "no HTTP"
  assertion would be a false green (dead oracle).

## Steps

1. Configure real-grok open harness (grouping).
2. Clear `NoSubmit`; rebuild args without `--no-submit`.
3. Run; settle; read log-http + scan GROK_HOME.
4. Assert turn evidence exists for the prompt.

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.NoSubmit = false
	req.Prompt = defaultOptionCPrompt
	req.Mode = "open-real-grok-after"
	req.OpenInstantAttach = true
	// Slightly longer settle: control needs a completed chat round-trip.
	if req.SettleAfter < 8*time.Second {
		req.SettleAfter = 8 * time.Second
	}
	req.Args = openRealGrokArgs(req)
	return nil
}
```
