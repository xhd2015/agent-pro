# Scenario

**Feature**: multi-attach write/observer policy on live stub-tty sessions

```
first attach_mode=interactive -> unified writer; second+ -> permanent observer
server tty send -> WriteInput (never client write slot)
```

## Preconditions

- Background `stub-tty --keep-tty` session with keep-alive scenario.
- ptywrap supports `attach_mode` interactive | observer | snapshot.

## Steps

1. Start background stub-tty with `--keep-tty`.
2. Leaf sets `req.Operation = "multi-attach"` and `req.Action`.
3. `Run` probes writer/observer/send behavior via WebSocket + CLI.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Operation = "multi-attach"
	req.EnableStubTTY = true
	req.StubScenarioJSON = stubScenarioKeepAliveJSON()
	return nil
}
```
