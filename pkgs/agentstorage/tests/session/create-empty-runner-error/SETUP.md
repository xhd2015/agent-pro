# Scenario

**Feature**: CreateSession rejects empty meta.Runner

```
CreateSession(id, meta.Runner="") -> error; no session directory created
```

## Preconditions

- `meta.Runner` is required metadata for create.
- Empty runner must not create a partial session dir.

## Steps

1. Set `req.Action = "create_empty_runner"`.
2. Set session id; leave runner empty via action.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Action = "create_empty_runner"
	req.SessionID = "sess_no_runner"
	return nil
}
```
