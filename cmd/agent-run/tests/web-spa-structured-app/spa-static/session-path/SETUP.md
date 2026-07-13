# Scenario

**Feature**: `/sessions/:sessionId` SPA shell — bootstrap depends on store hit

```
# unknown id: SPA shell without bootstrap
GET /sessions/no-such -> 200 HTML #root, no agent-run-session-bootstrap

# seeded id: shell + bootstrap JSON with session_id
seed sessions/<id> -> GET /sessions/<id> -> #agent-run-session-bootstrap
```

## Preconditions

- Path shape is exactly two segments: `sessions` + non-empty id.
- Store is flat: `AGENT_RUN_HOME/sessions/<session_id>/`.

## Steps

1. Narrow scenario under session-path.
2. Leaves seed or skip seed, then HTTP GET the session page path.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Grouping marker — leaves set SessionID and seed policy.
	if req.Scenario == "" {
		req.Scenario = "session-path"
	}
	return nil
}
```
