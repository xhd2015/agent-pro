# Scenario

**Feature**: `kill` argument validation — missing and unknown session ids

```
agent-run kill -> error, non-zero (missing session-id)
agent-run kill no-such-id -> error, non-zero (not found / expired)
```

## Preconditions

- Empty `AGENT_RUN_HOME` (no registry / no sessions).
- Mode handle.

## Steps

1. Leaf sets Args for the validation case.
2. Run Handle.
3. Assert non-zero exit and error text.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	req.Mode = "handle"
	return nil
}
```
