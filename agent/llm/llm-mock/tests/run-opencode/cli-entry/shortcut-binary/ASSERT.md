---
label: e2e
---

## Expected

- `llm-mock-run-opencode` exits 0 with fake opencode hook.
- Output contains `OPENCODE_CONFIG_DIR=` confirming orchestrator ran.

## Exit Code

0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if req.ShortcutPath == "" {
		t.Fatal("shortcut binary not built: implement agent/llm/llm-mock/llm-mock-run-opencode")
	}
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout+resp.Stderr, "OPENCODE_CONFIG_DIR=")
}
```