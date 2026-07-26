---
label: e2e
---

## Expected

- HTTP **200** on session detail.
- `session.workspace` equals the web server's cwd (`req.TempDir`).

```go
import (
	"encoding/json"
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.HTTPStatus != 200 {
		t.Fatalf("expected HTTP 200, got %d body=%q", resp.HTTPStatus, resp.HTTPBody)
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(resp.HTTPBody), &parsed); err != nil {
		t.Fatalf("parse detail JSON: %v", err)
	}
	sess, _ := parsed["session"].(map[string]any)
	if sess == nil {
		t.Fatalf("missing session object: %q", resp.HTTPBody)
	}
	workspace, _ := sess["workspace"].(string)
	want := filepath.Clean(req.TempDir)
	got := filepath.Clean(workspace)
	if got != want {
		t.Fatalf("session.workspace = %q, want %q (server cwd)", got, want)
	}
}
```