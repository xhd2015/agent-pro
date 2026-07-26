---
label: e2e
---

## Expected

- HTTP **200**.
- Response body includes `workspace` (or selected) equal to `SelectPath`.
- `config.json`: `selected_workspace` = SelectPath; `recent_workspaces[0]` = SelectPath.

## Errors

- Pre-impl: 404/405 on PUT (RED).

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
		t.Fatalf("PUT expected 200, got %d body=%q", put.Status, truncate(put.Body, 300))
	}
	m := parseJSONMap(t, put.Body)
	ws := jsonStringField(m, "workspace")
	if ws == "" {
		ws = jsonStringField(m, "selected_workspace")
	}
	if !pathsEqual(ws, req.SelectPath) {
		t.Fatalf("PUT body workspace: got %q want %q", ws, req.SelectPath)
	}
	recent := jsonStringSlice(m, "recent_workspaces")
	if len(recent) < 1 || !pathsEqual(recent[0], req.SelectPath) {
		t.Fatalf("recent_workspaces head: got %v want [%q, …]", recent, req.SelectPath)
	}
	cfg := readHomeConfigMap(t, req.Home)
	if cfg == nil {
		t.Fatal("expected config.json written after PUT")
	}
	sel, _ := cfg["selected_workspace"].(string)
	if !pathsEqual(sel, req.SelectPath) {
		t.Fatalf("config selected_workspace: got %q want %q", sel, req.SelectPath)
	}
	cfgRecent := jsonStringSlice(cfg, "recent_workspaces")
	if len(cfgRecent) < 1 || !pathsEqual(cfgRecent[0], req.SelectPath) {
		t.Fatalf("config recent_workspaces: got %v want head %q", cfgRecent, req.SelectPath)
	}
}
```
