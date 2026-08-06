# Scenario

**Feature**: Status errors when session id is not found

```
empty sessions tree (no summary for id)
-> Status(unknown-id, ...)
-> error "grok session not found: <id>"
```

## Preconditions

- Grok home has `sessions/` but no matching session directory for the id.
- Active list and live injectables are irrelevant once Find fails.

## Steps

1. Do not write a session for the id.
2. Set `SessionID` to a known-missing UUID.
3. Optional empty active list.

```go
import "testing"

const unknownStatusSessionID = "019f283a-eeee-7eee-eeee-eeeeeeeeee99"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = unknownStatusSessionID
	writeActiveSessions(t, req.GrokHome /* none */)
	req.Procs = nil
	req.OpenFiles = map[int][]string{}
	return nil
}
```
