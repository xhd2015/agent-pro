## Expected

- Exit code 0.
- Session `meta.workspace` equals the process cwd (`req.TempDir`), not a sibling directory.

## Exit Code

0

```go
import (
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)

	cwd := processCwd(req)
	assertSessionWorkspace(t, req.Home, cwd)

	got := sessionWorkspace(t, req.Home)
	other := filepath.Join(req.TempDir, "other-ws")
	if pathsEqual(got, other) {
		t.Fatalf("without --dir, workspace must not be sibling %q; got %q", other, got)
	}
}
```
