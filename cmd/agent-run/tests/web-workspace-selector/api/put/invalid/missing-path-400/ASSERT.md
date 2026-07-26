---
label: e2e
---

## Expected

- HTTP **400**.
- `config.json` `selected_workspace` still equals baseline (`req.SelectPath`).

## Errors

- Pre-impl: route missing may yield 404 (still RED vs 400 contract).

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
		t.Fatalf("PUT missing path expected 400, got %d body=%q", put.Status, truncate(put.Body, 300))
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
