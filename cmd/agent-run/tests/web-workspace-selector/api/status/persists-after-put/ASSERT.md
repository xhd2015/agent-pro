---
label: e2e
---

## Expected

- PUT returns **200**.
- Subsequent GET status returns `workspace` equal to `SelectPath`.
- `config.json` contains `selected_workspace` = `SelectPath`.

## Errors

- Pre-impl: PUT route missing → non-200 (RED).

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
	if put.Status != 200 {
		t.Fatalf("PUT workspace expected 200, got %d body=%q", put.Status, truncate(put.Body, 300))
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
	if !pathsEqual(ws, req.SelectPath) {
		t.Fatalf("status.workspace after PUT: got %q want %q", ws, req.SelectPath)
	}
	cfg := readHomeConfigMap(t, req.Home)
	if cfg == nil {
		t.Fatal("expected config.json after PUT")
	}
	sel, _ := cfg["selected_workspace"].(string)
	if !pathsEqual(sel, req.SelectPath) {
		t.Fatalf("config selected_workspace: got %q want %q", sel, req.SelectPath)
	}
}
```
