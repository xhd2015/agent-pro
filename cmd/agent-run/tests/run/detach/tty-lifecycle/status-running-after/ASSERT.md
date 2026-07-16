## Expected

- Exit code 0.
- Both ids on stdout.
- `meta.status` is `running` (parent did not wait for full turn / finished).
- Registry listen_addr TCP-reachable.

## Side Effects

- Session remains attachable/sendable after parent exits.

## Exit Code

0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v\nstdout:\n%s\nstderr:\n%s", err, resp.Stdout, resp.Stderr)
	}
	assertSuccess(t, resp)
	sid, tid := assertDetachIDsOnStdout(t, resp)

	meta := resp.MetaAfter
	if meta == nil {
		if !fileExists(metaJSONPath(req.Home, sid)) {
			t.Fatalf("meta.json missing for session %s after detach", sid)
		}
		meta = readMetaJSON(t, req.Home, sid)
	}
	status, _ := meta["status"].(string)
	if strings.TrimSpace(status) != "running" {
		t.Fatalf("meta.status want running, got %q (meta=%v)\nstderr:\n%s", status, meta, resp.Stderr)
	}

	// Terminal reachable via registry.
	entry := resp.RegistryEntry
	if entry == nil {
		var rerr error
		entry, rerr = readRegistryEntryOptional(req.Home, "grok-tty", tid)
		if rerr != nil && sid != tid {
			entry, rerr = readRegistryEntryOptional(req.Home, "grok-tty", sid)
		}
		if rerr != nil {
			t.Fatalf("registry missing after detach: %v", rerr)
		}
	}
	if entry.ListenAddr == "" || !portOpen(entry.ListenAddr) {
		t.Fatalf("terminal not reachable after detach; listen_addr=%q", entry.ListenAddr)
	}
}
```
