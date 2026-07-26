---
label: e2e
---

## Expected

- Exit code 0.
- Stderr contains `grok-tty: <id>` where `<id>` matches auto-id shape with base
  `hello-world`.
- Storage dir `sessions/grok-tty/<id>/` exists.
- `meta.session_id` and `meta.terminal_session_id` both equal `<id>`.
- Registry file `grok-tty-registry/<id>.json` exists (kept via `--keep-tty`).

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
		t.Fatal(err)
	}
	assertSuccess(t, resp)

	stderrID, ok := parseGrokTTYSessionID(resp.Stderr)
	if !ok {
		t.Fatalf("missing grok-tty session id on stderr:\n%s", resp.Stderr)
	}
	base, _, _, ok := splitAutoSessionID(stderrID)
	if !ok {
		t.Fatalf("stderr id %q does not match auto-id shape", stderrID)
	}
	if base != "hello-world" {
		t.Fatalf("stderr id base = %q, want hello-world (id=%q)", base, stderrID)
	}

	storageID := singleSessionID(t, req.Home, "grok-tty")
	if storageID != stderrID {
		t.Fatalf("storage id %q != stderr id %q", storageID, stderrID)
	}

	meta := readSessionMeta(t, req.Home, "grok-tty", storageID)
	if meta.SessionID != storageID {
		t.Fatalf("meta.session_id = %q, want %q", meta.SessionID, storageID)
	}
	if meta.TerminalSessionID != storageID {
		t.Fatalf("meta.terminal_session_id = %q, want same as storage %q", meta.TerminalSessionID, storageID)
	}

	regPath := grokTTYRegistryPath(req.Home, storageID)
	if _, err := os.Stat(regPath); err != nil {
		t.Fatalf("registry entry missing at %s: %v\nstderr:\n%s", regPath, err, resp.Stderr)
	}
}
```
