# Scenario

**Feature**: resolve with empty --grok-session-id value uses library empty-id error

```
RunSessions(["resolve", "--grok-session-id", ""])
  -> err: --grok-session-id requires a non-empty value
  -> stdout empty
```

## Steps

1. Flag present with empty string value (distinct from omitted).
2. No seeds required.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"resolve", "--grok-session-id", ""}
	return nil
}
```
