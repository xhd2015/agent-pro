# Scenario

**Feature**: Attach bridges local TTY to remote WS session

```
# attach bridge
local TTY <-> WS <-> daemon PTY session
```

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "attach-requires-tty"
	return nil
}
```