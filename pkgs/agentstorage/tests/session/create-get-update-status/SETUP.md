# Scenario

**Feature**: session create, get, and status update roundtrip

```
CreateSession(running) -> GetSession -> UpdateSessionStatus(finished) -> GetSession
```

## Preconditions

- Session does not exist before create.
- Status transitions from `running` to `finished`.

## Steps

1. Set `req.Action = "create_get_update"`.
2. Set session id, runner session id, model, and target status.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Action = "create_get_update"
	req.SessionID = "sess_crud"
	req.RunnerSess = "runner-sess-1"
	req.Model = "test-model"
	req.Status = "finished"
	return nil
}
```