---
label: e2e
---

## Expected

- HTTP **200**.
- `workspace` equals process cwd (`WebWorkingDir`).
- `process_cwd` equals `WebWorkingDir` (new field for Quick Server cwd).
- `home` is present and equals OS user home (Quick Home chip source).
- `recent_workspaces` is empty array or absent/empty.

## Side Effects

- No `selected_workspace` written solely by GET status.

## Errors

- Partial today: `workspace` may already equal cwd; missing `process_cwd` /
  `recent_workspaces` / user-home semantics → RED until status is extended.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	st, ok := findHTTPResult(resp, "status")
	if !ok {
		t.Fatal("missing status result")
	}
	if st.Status != 200 {
		t.Fatalf("GET status expected 200, got %d body=%q", st.Status, truncate(st.Body, 300))
	}
	m := parseJSONMap(t, st.Body)
	ws := jsonStringField(m, "workspace")
	if !pathsEqual(ws, req.WebWorkingDir) {
		t.Fatalf("workspace: got %q want %q (process cwd)", ws, req.WebWorkingDir)
	}
	proc := jsonStringField(m, "process_cwd")
	if proc == "" {
		t.Fatal("expected process_cwd on status (Quick Server cwd source)")
	}
	if !pathsEqual(proc, req.WebWorkingDir) {
		t.Fatalf("process_cwd: got %q want %q", proc, req.WebWorkingDir)
	}
	home := jsonStringField(m, "home")
	if home == "" {
		t.Fatal("expected home on status (Quick Home source)")
	}
	if req.OSUserHome != "" && !pathsEqual(home, req.OSUserHome) {
		// store.Home() is wrong semantics for Quick Home chip
		t.Fatalf("home: got %q want OS user home %q", home, req.OSUserHome)
	}
	recent := jsonStringSlice(m, "recent_workspaces")
	if len(recent) != 0 {
		t.Fatalf("fresh recent_workspaces must be empty, got %v", recent)
	}
}
```
