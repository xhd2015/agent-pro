# Scenario

**Feature**: screen status shows "unknown" when ptywrap server is unreachable

```
agent-run tty status session-1 -> registry with reachable listen addr but server dead -> screen: unknown
```

## Steps

1. Registry entry has a listen_addr pointing to a port with no real server.
2. `req.Args` = `["tty", "status", "session-1"]`.
3. `Assert` checks that screen status is "unknown" or error indicator.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.StartFakePTYWrap = false
	req.FakePTYWrapPort = 0
	req.RegistryEntryJSON = defaultRegistryEntryJSON(req.RegistrySessionID, 12345)
	writeMockRegistryEntry(t, req)
	req.Args = []string{"tty", "status", req.RegistrySessionID}
	return nil
}
```
