# Scenario

**Feature**: `tty send` fails when registry entry exists but port is unreachable

```
agent-run tty send session-1 "hello" -> registry with dead port -> error
```

## Steps

1. Registry entry has listen_addr on a closed port.
2. `req.Args` = `["tty", "send", "session-1", "hello"]`.
3. `req.Mode` = `"send-probe"`.
4. Expect exit code 1 or error output.

```go
import (
	"github.com/xhd2015/doctest/session"
	"time"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"tty", "send", req.RegistrySessionID, "hello"}
	req.Mode = "send-probe"
	req.ExecTimeout = 15 * time.Second
	return nil
}
```
