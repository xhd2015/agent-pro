# Scenario

**Feature**: session history human lines oldest→newest

```
session history --session-id ID -> three human lines chronological
```

## Steps

1. Default human output.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"session", "history",
		"--session-id", sessionHistoryFixtureID,
	}
	return nil
}
```
