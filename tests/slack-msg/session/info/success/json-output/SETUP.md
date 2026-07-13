# Scenario

**Feature**: session info --json

```
session info --session-id ID --json -> JSON object with message_count + session_dir
```

## Steps

1. Pass --json.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"session", "info",
		"--session-id", sessionInfoFixtureID,
		"--json",
	}
	return nil
}
```
