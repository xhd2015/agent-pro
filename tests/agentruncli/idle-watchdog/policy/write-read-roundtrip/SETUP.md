# Scenario

**Feature**: write then read `exit_on_idle=true`, `idle_timeout=10m`

```
WriteIdlePolicy(home, sess, {true, 10m})
  -> file at sessions/<sess>/idle-policy.json
ReadIdlePolicy -> found, same fields, compact "10m" in JSON
```

## Steps

1. Write `ExitOnIdle=true`, `IdleTimeout=10m`.
2. Read back; do not Tick.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.WritePolicy = true
	req.ExitOnIdle = true
	req.IdleTimeout = 10 * time.Minute
	req.SessionID = "sess-policy-roundtrip"
	return nil
}
```
