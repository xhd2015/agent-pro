---
label: e2e
---

## Expected

- Exit code 0.
- Both ids on stdout.
- Registry file exists for terminal-id (or equal session-id).
- listen_addr TCP-reachable.

## Exit Code

0

```go
import (
	"os"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v\nstdout:\n%s\nstderr:\n%s", err, resp.Stdout, resp.Stderr)
	}
	assertSuccess(t, resp)
	_, tid := assertDetachIDsOnStdout(t, resp)

	regPath := registryPath(req.Home, "grok-tty", tid)
	if _, statErr := os.Stat(regPath); statErr != nil {
		if resp.SessionID != "" && resp.SessionID != tid {
			alt := registryPath(req.Home, "grok-tty", resp.SessionID)
			if _, e2 := os.Stat(alt); e2 == nil {
				regPath = alt
				tid = resp.SessionID
			} else {
				t.Fatalf("registry missing after resume --detach at %s: %v", regPath, statErr)
			}
		} else {
			t.Fatalf("registry missing after resume --detach at %s: %v", regPath, statErr)
		}
	}
	entry := resp.RegistryEntry
	if entry == nil {
		var rerr error
		entry, rerr = readRegistryEntryOptional(req.Home, "grok-tty", tid)
		if rerr != nil {
			t.Fatalf("read registry: %v", rerr)
		}
	}
	if entry.ListenAddr == "" || !portOpen(entry.ListenAddr) {
		t.Fatalf("listen_addr %q not reachable after resume --detach", entry.ListenAddr)
	}
}
```
