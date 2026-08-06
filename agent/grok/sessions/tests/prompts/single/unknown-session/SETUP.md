# Scenario

**Feature**: unknown session id returns not-found error

```
Prompts(grokHome, unknownID) -> error "grok session not found: <id>"
```

## Preconditions

- No session directory for `idUnknown`.
- Grok home has empty `sessions/` tree only.

## Steps

1. Set SessionID to unknown id; do not write a session.
2. Call Prompts.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "single"
	req.SessionID = idUnknown
	return nil
}
```
