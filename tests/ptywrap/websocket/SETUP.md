# Scenario

**Feature**: WebSocket attach protocol for PTY I/O and control

```
# WS protocol
client -> binary keystrokes / JSON resize -> PTY
PTY output -> binary frames + scrollback replay on reconnect
```

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.AttachMode = ""
	return nil
}
```