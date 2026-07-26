# Scenario

**Feature**: writer normal WS close must free the OS PTY (no orphan shell)

When a terminal client disconnects with WebSocket close code **1000** (normal
closure — browser refresh, tab unmount, LocalTerminal cleanup), the server must
reap the session child so the kernel PTY is released.

Keeping the in-memory session for scrollback replay is fine; keeping a live
`sleep`/`bash` process is the system-wide `openpty: Device not configured` leak.

```
# leak path (current)
WS attach as writer -> close 1000 -> child still running + holds PTY

# expected
WS attach as writer -> close 1000 -> ProcessAlive == false
```

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "lifecycle-writer-close"
	req.WSCloseCode = 1000
	return nil
}
```
