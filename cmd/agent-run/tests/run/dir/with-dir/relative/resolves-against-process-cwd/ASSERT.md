---
label: e2e
---

## Expected

- Exit code 0.
- `meta.workspace` equals the absolute path of `TempDir/rel-ws` (resolved against process cwd).

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
	want := filepath.Join(req.TempDir, "rel-ws")
	assertSessionWorkspace(t, req.Home, want)
}
```
