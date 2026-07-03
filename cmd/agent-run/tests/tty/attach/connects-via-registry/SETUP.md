# Scenario

**Feature**: `tty attach` connects successfully via registry entry

```
agent-run tty attach session-1 -> registry lookup -> ptyclient.Attach -> connection established
```

## Steps

1. Start fake ptywrap WS server, write registry entry pointing to it.
2. `req.Args` = `["tty", "attach", "session-1"]`.
3. `req.Mode` = `"attach-probe"` for timeout-based probe.
4. Since attach blocks interactively, the probe verifies the WS handshake succeeds.

```go
import "fmt"

func Setup(t *testing.T, req *Request) error {
	req.StartFakePTYWrap = true
	startFakePTYWrapServer(t, req)
	waitForPortOpen(t, fmt.Sprintf("127.0.0.1:%d", req.FakePTYWrapPort), 5*time.Second)

	req.RegistryDir = "grok-tty-registry"
	req.RegistrySessionID = "session-1"
	req.RegistryEntryJSON = defaultRegistryEntryJSON(req.RegistrySessionID, req.FakePTYWrapPort)
	writeMockRegistryEntry(t, req)

	req.Args = []string{"tty", "attach", req.RegistrySessionID}
	req.Mode = "attach-probe"
	req.AttachTimeout = 5 * time.Second
	return nil
}
```
