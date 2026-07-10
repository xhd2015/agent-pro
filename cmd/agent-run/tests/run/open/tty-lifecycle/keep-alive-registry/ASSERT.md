## Expected

- Exit code 0.
- Session id present on stderr (`grok-tty: <id>`).
- Registry file `AGENT_RUN_HOME/grok-tty-registry/<id>.json` exists after open.
- Registry `listen_addr` is non-empty and TCP-reachable (session still alive).

## Side Effects

- Keep-alive TTY session remains after the open command returns (for re-attach/send).

## Exit Code

0

```go
import (
	"os"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v\nstdout:\n%s\nstderr:\n%s", err, resp.Stdout, resp.Stderr)
	}
	assertSuccess(t, resp)

	id, ok := parsePrefixedSessionID(resp.Stderr, "grok-tty")
	if !ok {
		t.Fatalf("missing grok-tty session id on stderr:\n%s", resp.Stderr)
	}

	regPath := registryPath(req.Home, "grok-tty", id)
	if _, statErr := os.Stat(regPath); statErr != nil {
		t.Fatalf("registry entry missing after --open keep-alive at %s: %v\nstderr:\n%s",
			regPath, statErr, resp.Stderr)
	}

	entry := resp.RegistryEntry
	if entry == nil {
		var rerr error
		entry, rerr = readRegistryEntryOptional(req.Home, "grok-tty", id)
		if rerr != nil {
			t.Fatalf("read registry %s: %v", regPath, rerr)
		}
	}
	if entry.ListenAddr == "" {
		t.Fatalf("registry entry has empty listen_addr for session %s", id)
	}
	if !portOpen(entry.ListenAddr) {
		t.Fatalf("listen_addr %s not reachable after --open (keep-alive required)", entry.ListenAddr)
	}
}
```
