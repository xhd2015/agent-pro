# Scenario

**Bug**: `agent-run run --open --no-submit "draft"` still submits the draft as
the first model turn when using real Grok (positional argv on new open sessions)

```
agent-run run --agent-runner grok-tty --open --no-submit "draft-no-submit-OPTIONC-probe-zz9"
  --agent-runner-binary llm-mock-run-grok
  --agent-runner-config-home <isolated>
  -> exit 0
  -> stderr once "grok-tty: <terminal-id>"
  -> open bind soft-unbound OK (must not hard-fail solely for missing session)
  -> after settle: NO mock HTTP chat for draft
  -> after settle: NO session user_message_chunk for draft
```

## Preconditions

- Option C: real `grok` under `llm-mock-run-grok`; **no**
  `LLM_MOCK_RUN_GROK_COMMAND`; **no** `AGENT_RUN_GROK_TTY_COMMAND`.
- Primary oracle is turn absence (HTTP / session), not a fake `SUBMITTED:` marker.
- Soft unbound: exit must not fail solely as `grok session id not resolved`
  because the draft was not submitted.

## Steps

1. Configure real-grok open harness (grouping).
2. Force `--open --no-submit` + distinctive draft prompt.
3. Run; settle; read log-http + scan GROK_HOME.
4. Assert exit 0, terminal id, and **no** turn evidence for the draft.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.NoSubmit = true
	req.Prompt = defaultOptionCPrompt
	req.Mode = "open-real-grok-after"
	req.OpenInstantAttach = true
	req.Args = openRealGrokArgs(req)
	return nil
}
```
