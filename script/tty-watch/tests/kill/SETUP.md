# Scenario

**Feature**: `tty-watch kill` terminates sessions and prunes registry

```
# kill detached session
tty-watch kill <id> -> DELETE ptywrap session -> SIGTERM owner -> remove registry json
```

## Preconditions

- `terminates-detached` starts a live detached session before kill.
- `prunes-unreachable` seeds a stale registry entry with dead listen addr.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.Bin == "" {
		t.Fatalf("kill setup: tty-watch binary not built")
	}
	return nil
}
```