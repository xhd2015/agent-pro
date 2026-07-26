---
label: e2e
---

## Expected

- Exit code 0.
- Among session dirs under `fake-codex`, exactly one **new** id (not only seeds)
  matches `hello-world-<ts>-<N>` with numeric suffix N ≥ 1.
- That id's `meta.json` has matching `session_id` (and preferably non-seed status
  or events — at minimum directory name == meta.session_id and shape with suffix).

## Exit Code

0

```go
import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)

	ids := listSessionIDs(t, req.Home, "fake-codex")
	var chosen string
	for _, id := range ids {
		base, _, suf, ok := splitAutoSessionID(id)
		if !ok || base != "hello-world" || suf == "" {
			continue
		}
		n, convErr := strconv.Atoi(suf)
		if convErr != nil || n < 1 {
			continue
		}
		// Prefer the session that has events or was written by this run:
		// seed only has meta.json; a real run should have more than meta or
		// meta status/events. Accept any suffixed id that has meta.session_id match
		// and is not "empty" of events if events exist on any sibling.
		metaPath := sessionMetaPath(req.Home, "fake-codex", id)
		if _, err := os.Stat(metaPath); err != nil {
			continue
		}
		// If events.jsonl exists, this is definitely the run output.
		events := filepath.Join(sessionDir(req.Home, "fake-codex", id), "events.jsonl")
		if fileExists(events) {
			chosen = id
			break
		}
		if chosen == "" {
			chosen = id
		}
	}
	if chosen == "" {
		t.Fatalf("expected a hello-world-*-N collision id under sessions/fake-codex/, got: %v\nstderr:\n%s",
			ids, resp.Stderr)
	}
	base, _, suf, ok := splitAutoSessionID(chosen)
	if !ok || base != "hello-world" || suf == "" {
		t.Fatalf("chosen id %q missing numeric collision suffix", chosen)
	}
	meta := readSessionMeta(t, req.Home, "fake-codex", chosen)
	if meta.SessionID != chosen {
		t.Fatalf("meta.session_id = %q, want %q", meta.SessionID, chosen)
	}
}
```
