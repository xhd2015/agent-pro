---
label: e2e
---

## Expected

- Exit code 0.
- Stdout both `session-id:` and `terminal-id:` (session-id may equal the
  explicit `--session-id`).
- Meta created; status `running`.
- No stream noise; not a live send `msg_N` path.

## Exit Code

0

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v\nstdout:\n%s\nstderr:\n%s", err, resp.Stdout, resp.Stderr)
	}
	assertNotUnknownFlag(t, strings.ToLower(resp.Stderr+"\n"+resp.Stdout), "--detach")
	assertSuccess(t, resp)
	sid, _ := assertDetachIDsOnStdout(t, resp)
	if req.SessionID != "" && sid != req.SessionID {
		// explicit --session-id should surface as session-id line
		t.Fatalf("session-id want %q got %q\nstdout:\n%s", req.SessionID, sid, resp.Stdout)
	}
	if noise := forbiddenDetachNoise(resp.Stdout + "\n" + resp.Stderr); len(noise) > 0 {
		t.Fatalf("auto MODE=run detach must not stream noise %v", noise)
	}
	first := strings.TrimSpace(strings.Split(resp.Stdout, "\n")[0])
	if strings.HasPrefix(first, "msg_") {
		t.Fatalf("MODE=run detach must not take send path; stdout:\n%s", resp.Stdout)
	}

	meta := resp.MetaAfter
	if meta == nil {
		if !fileExists(metaJSONPath(req.Home, sid)) {
			t.Fatalf("meta.json missing after auto detach create at %s", metaJSONPath(req.Home, sid))
		}
		meta = readMetaJSON(t, req.Home, sid)
	}
	status, _ := meta["status"].(string)
	if strings.TrimSpace(status) != "running" {
		t.Fatalf("meta.status want running, got %q", status)
	}
}
```
