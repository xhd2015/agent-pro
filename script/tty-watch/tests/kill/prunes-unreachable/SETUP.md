# Scenario

**Feature**: kill on unreachable server prunes stale registry idempotently

```
# stale registry with dead listen addr
harness seeds session-stale-1 -> tty-watch kill -> exit 0, registry pruned
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "kill-stale"
	return nil
}
```