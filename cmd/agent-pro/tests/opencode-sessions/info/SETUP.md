# Scenario

**Feature**: detailed info for a single OpenCode session

```
# locate session JSON by exact id
sessions.Find(dataDir, sessionID) -> session path

# aggregate session fields, message dir, token/cost totals
sessions.Info -> FormatInfoText(now) -> key-value output
```

## Preconditions

- This branch tests the `info` operation.
- Session id must be an exact match (no prefix search).

## Steps

1. Set `req.Operation = "info"`.
2. Leaf Setup writes session JSON and optional message files for `req.SessionID`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Operation = "info"
	return nil
}
```