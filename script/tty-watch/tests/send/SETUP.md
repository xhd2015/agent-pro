# Scenario

**Feature**: `tty-watch send` injects follow-up bytes verbatim into a live PTY session

```
# registry lookup then HTTP inject (no suffix, no trim)
tty-watch send <id> <message...> -> ReadRegistry -> POST /api/terminal/sessions/{id}/input -> exit 0 (silent)
```

## Preconditions

- Success leaves start a detached byte-capture child (`cat > capture.bin`) before send.
- `terminal-unreachable` seeds a stale registry entry with dead listen addr (like snapshot).
- Message bytes are joined from CLI args after session id with a single space.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.Bin == "" {
		t.Fatalf("send setup: tty-watch binary not built")
	}
	return nil
}
```