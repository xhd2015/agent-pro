# Scenario

**Feature**: sessions-level help indexes the resolve verb

```
RunSessions(["-h"])
  -> err nil
  -> stdout mentions resolve (and --grok-session-id)
```

## Steps

1. Args `-h` only (sessions help via injected API).
2. Assert stdout contains `resolve`.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"-h"}
	return nil
}
```
