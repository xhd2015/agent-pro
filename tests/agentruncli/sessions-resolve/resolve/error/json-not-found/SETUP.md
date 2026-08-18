# Scenario

**Feature**: --json with missing UUID still returns error (no JSON error body)

```
RunSessions(["resolve", "--json", "--grok-session-id", missing])
  -> not-found error; stdout empty or non-JSON
```

## Steps

1. No matching seeds.
2. Resolve with `--json` and unknown UUID.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"resolve", "--json", "--grok-session-id", "99999999-9999-9999-9999-999999999999"}
	return nil
}
```
