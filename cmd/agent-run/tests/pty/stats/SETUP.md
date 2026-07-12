# Scenario

**Feature**: `agent-run pty stats` reports PTY and agent-run serve summary

```
agent-run pty stats
  -> limit / unique masters / free estimate (best-effort)
  -> agent-run __serve count and category breakdown
  -> exit 0; stdout ends with \n
```

## Steps

1. Leaf `Setup` sets `req.Args` to `pty stats`.
2. `Run` executes stats (read-only; may see host serves).
3. `Assert` checks exit 0, summary keywords, trailing newline.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if len(req.Args) == 0 {
		req.Args = []string{"pty", "stats"}
	}
	return nil
}
```
