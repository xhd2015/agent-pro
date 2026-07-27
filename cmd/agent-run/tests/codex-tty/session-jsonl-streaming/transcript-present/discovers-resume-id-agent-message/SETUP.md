# Scenario

**Feature**: resume UUID discovers rollout JSONL and streams `event_msg.agent_message`

```
scrollback: codex resume <uuid>
  -> CODEX_HOME rollout-*-<uuid>.jsonl
  -> event_msg.agent_message.message
  -> assistant stdout
```

## Preconditions

- The assistant text appears only in the rollout JSONL, not in fake TUI scrollback.

## Steps

1. Seed the matching rollout JSONL with `event_msg.agent_message`.
2. Assert the seeded assistant message reaches stdout.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

const codexAgentMessageText = "JSONL_DISCOVERED_AGENT_MESSAGE"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	seedCodexTranscript(t, req, codexAgentMessageLine(codexAgentMessageText))
	req.StreamProbeSubstring = codexAgentMessageText
	return nil
}
```
