# Scenario

**Feature**: CreateSession rejects duplicate bare session id

```
CreateSession(id) succeeds -> CreateSession(same id) -> error; first meta not overwritten
```

## Preconditions

- Session directory already exists for the id.
- Second create must not wipe or rewrite existing meta.

## Steps

1. Set `req.Action = "create_duplicate"`.
2. Use a fixed session id.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "create_duplicate"
	req.SessionID = "sess_dup"
	return nil
}
```
