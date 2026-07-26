# Scenario

**Feature**: get missing session returns error

```
GetSession(unknown_id) -> non-nil error, nil session
```

## Preconditions

- No session directory exists for the requested id.
- Store has been opened on a fresh home.

## Steps

1. Set `req.Action = "get_missing"`.
2. Request a session id that was never created.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "get_missing"
	req.SessionID = "sess_does_not_exist"
	return nil
}
```
