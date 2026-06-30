# Scenario

**Feature**: session detail API exposes timeline `message` events with `role` and `timestamp`

```
POST /api/agent-run/sessions {prompt} -> user message event (timestamp ms) -> GET detail
agentui.Run emit -> assistant message event (timestamp ms) -> GET detail
POST .../messages {text} -> user follow-up event -> GET detail
continuation: two-turn flow -> assistant recalls first user text
streaming: agentui.Run -> assistant message events with phase start|update|end
```

## Preconditions

- Web server uses explicit Bearer token (`test`).
- `fake-codex` runner is on `PATH` for create-session flows that start an agent run.
- User timeline entries use `type=message` and `role=user`.

## Steps

1. Leaf sets `req.Mode = "web"` and starts `agent-run web` in the background.
2. Leaf performs HTTP setup (create session and/or follow-up message) before `Run`.
3. `Run` GETs session detail; `Assert` polls until user event appears when needed.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "web"
	return nil
}
```