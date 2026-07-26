---
label: e2e
---

## Expected

- HTTP **400** (must be a directory).
- Baseline `selected_workspace` unchanged.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	put, ok := findHTTPResult(resp, "put")
	if !ok {
		t.Fatal("missing put result")
	}
	if put.Status != 400 {
		t.Fatalf("PUT file path expected 400, got %d body=%q", put.Status, truncate(put.Body, 300))
	}
	cfg := readHomeConfigMap(t, req.Home)
	if cfg == nil {
		t.Fatal("expected baseline config.json to remain")
	}
	sel, _ := cfg["selected_workspace"].(string)
	if !pathsEqual(sel, req.SelectPath) {
		t.Fatalf("selected_workspace clobbered: got %q want baseline %q", sel, req.SelectPath)
	}
}
```
