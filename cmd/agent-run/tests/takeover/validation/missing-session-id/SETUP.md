# Scenario

**Feature**: `takeover` without a session-id fails

```
agent-run takeover -> requires <session-id>; exit non-zero
```

## Steps

1. Args = `["takeover"]` only.
2. Run Mode handle.
3. Assert non-zero exit and error mentions session / usage.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = "handle"
	req.Args = []string{"takeover"}
	return nil
}
```
