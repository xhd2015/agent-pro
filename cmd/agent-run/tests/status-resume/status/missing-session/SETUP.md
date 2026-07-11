# Scenario

**Feature**: status on unknown session exits 1 with not-found error

```
agent-run status no-such-session -> exit 1, not found
```

## Steps

1. Run status with an id that has no meta under any runner.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "no-such-session-xyz"
	req.Args = []string{"status", req.SessionID}
	return nil
}
```
