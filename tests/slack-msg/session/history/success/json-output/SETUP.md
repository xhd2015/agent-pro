# Scenario

**Feature**: session history --json

```
session history --session-id ID --json -> JSON document chronological
```

## Steps

1. Pass --json.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"session", "history",
		"--session-id", sessionHistoryFixtureID,
		"--json",
	}
	return nil
}
```
