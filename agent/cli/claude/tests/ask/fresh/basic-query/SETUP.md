# Scenario

**Feature**: Fresh Ask() with the default model returns the requested word and a session id

```
# default-model query: claude streams assistant text "pong" + result with session_id
basic-query -> ClaudeAgent.Ask("Reply with exactly the word: pong")
ClaudeAgent <- claude (assistant text "pong", result success session_id)
ClaudeEventWriter -> RawLog (AgentEvent JSONL with ≥1 "message")
```

## Preconditions
- The `claude` binary is available in PATH.
- This is a fresh session query (no model override, no session resume).

## Steps
1. Set the prompt to ask claude to reply with exactly the word "pong".

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Prompt = "Reply with exactly the word: pong"
	return nil
}
```
