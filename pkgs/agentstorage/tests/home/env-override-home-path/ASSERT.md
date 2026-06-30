## Expected

- `Store.Home()` equals the `AGENT_RUN_HOME` override path.
- Resolved home does not equal the constructor `req.Home` default.

```go
import (
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	want := filepath.Join(req.TempDir, "override-home")
	assertEqual(t, "ResolvedHome", resp.ResolvedHome, want)
	if resp.ResolvedHome == req.Home {
		t.Fatalf("expected env override %q to differ from constructor home %q", want, req.Home)
	}
}
```