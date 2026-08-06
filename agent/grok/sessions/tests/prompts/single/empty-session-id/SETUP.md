# Scenario

**Feature**: empty session id returns not-found error (same as Find)

```
Prompts(grokHome, "") -> error containing "grok session not found"
```

## Preconditions

- SessionID is empty string.
- Matches existing `sessionNotFoundError` style for empty id.

## Steps

1. Set SessionID to `""`.
2. Call Prompts.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "single"
	req.SessionID = ""
	return nil
}
```
