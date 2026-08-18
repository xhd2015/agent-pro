# Scenario

**Feature**: resolve without --grok-session-id requires the flag

```
RunSessions(["resolve"])
  -> err: sessions resolve requires --grok-session-id
  -> stdout empty
```

## Steps

1. No seeds.
2. Args `resolve` only (flag omitted entirely).

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"resolve"}
	return nil
}
```
