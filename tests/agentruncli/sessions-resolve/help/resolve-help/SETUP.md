# Scenario

**Feature**: `sessions resolve --help` documents Usage and flags

```
RunSessions(["resolve", "--help"])
  -> err nil
  -> stdout: Usage + --grok-session-id + --json
```

## Steps

1. Args `resolve --help` (no seeds).
2. Assert exact help template (implementer matches).

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"resolve", "--help"}
	return nil
}
```
