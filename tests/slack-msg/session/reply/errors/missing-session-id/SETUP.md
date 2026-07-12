# Scenario

**Feature**: session reply without session id

```
session reply MESSAGE (no --session-id, no SLACK_MSG_SESSION_ID)
  -> session id required; exit 1
```

## Steps

1. Message only.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"session", "reply", "hello"}
	return nil
}
```
