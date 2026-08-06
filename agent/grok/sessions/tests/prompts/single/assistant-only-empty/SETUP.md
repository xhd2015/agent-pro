# Scenario

**Feature**: session with only assistant/tool updates returns empty prompts

```
# no user_message_chunk
assistant + tool_call + turn_completed
  -> Prompts -> UserPrompts empty, Err nil
```

## Preconditions

- Session exists with valid summary.json.
- updates.jsonl has no user_message_chunk lines.

## Steps

1. Write assistant/tool-only session.
2. Call Prompts.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "single"
	req.SessionID = idAssistantOnly
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idAssistantOnly,
		Title:        "assistant only",
		LastActiveAt: atFixed(-time.Hour),
		Updates: updatesJSONL(
			assistantChunk("I am ready"),
			toolCallPending("call_1", "ls"),
			turnCompleted(),
		),
	})
	return nil
}
```
