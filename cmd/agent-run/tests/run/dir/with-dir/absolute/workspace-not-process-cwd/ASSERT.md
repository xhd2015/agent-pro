---
label: e2e
---

## Expected

- Exit code 0.
- Session `meta.workspace` equals the absolute `--dir` path (or EvalSymlinks form).
- `meta.workspace` is **not** the process cwd when they differ.

## Exit Code

0

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)

	wantWS := filepath.Join(req.TempDir, "ws-target")
	assertSessionWorkspace(t, req.Home, wantWS)

	got := sessionWorkspace(t, req.Home)
	cwd := processCwd(req)
	if pathsEqual(got, cwd) {
		t.Fatalf("meta.workspace %q must differ from process cwd %q when --dir targets another directory",
			got, cwd)
	}
}
```
