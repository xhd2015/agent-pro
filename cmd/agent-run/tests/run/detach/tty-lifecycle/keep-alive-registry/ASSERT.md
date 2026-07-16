## Expected

- Exit code 0.
- Both ids present on stdout.
- Registry file `AGENT_RUN_HOME/grok-tty-registry/<terminal-id>.json` exists.
- Registry `listen_addr` is non-empty and TCP-reachable.

## Side Effects

- Keep-alive TTY session remains after the detach command returns.

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
	_, tid := assertDetachIDsOnStdout(t, resp)

	regPath := registryPath(req.Home, "grok-tty", tid)
	if _, statErr := os.Stat(regPath); statErr != nil {
		// also try session-id as registry key if equal-id policy
		if resp.SessionID != "" && resp.SessionID != tid {
			alt := registryPath(req.Home, "grok-tty", resp.SessionID)
			if _, e2 := os.Stat(alt); e2 == nil {
				regPath = alt
				tid = resp.SessionID
			} else {
				t.Fatalf("registry entry missing after --detach keep-alive at %s: %v\nstderr:\n%s",
					regPath, statErr, resp.Stderr)
			}
		} else {
			t.Fatalf("registry entry missing after --detach keep-alive at %s: %v\nstderr:\n%s",
				regPath, statErr, resp.Stderr)
		}
	}

	entry := resp.RegistryEntry
	if entry == nil {
		var rerr error
		entry, rerr = readRegistryEntryOptional(req.Home, "grok-tty", tid)
		if rerr != nil {
			t.Fatalf("read registry %s: %v", regPath, rerr)
		}
	}
	if entry.ListenAddr == "" {
		t.Fatalf("registry entry has empty listen_addr for terminal %s", tid)
	}
	if !portOpen(entry.ListenAddr) {
		t.Fatalf("listen_addr %s not reachable after --detach (keep-alive required)", entry.ListenAddr)
	}
}
```
