---
label: e2e
---

## Expected

- Argv probe log contains `MOCK_WRAPPER_INVOKED=1` (web used configured mock binary).
- Grok home contains `updates.jsonl` for seeded session uuid under configured `--grok-home`.
- Session detail HTTP 200 and events include assistant stream marker `WEB_MOCK_STREAM_MARKER`.

## Exit Code

0

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.HTTPStatus != 200 {
		t.Fatalf("HTTP status = %d body=%q", resp.HTTPStatus, resp.HTTPBody)
	}
	if !strings.Contains(resp.ArgvProbe, "MOCK_WRAPPER_INVOKED=1") {
		t.Fatalf("mock wrapper not invoked; probe log:\n%s", resp.ArgvProbe)
	}
	if !grokSessionUpdatesExists(req.GrokHome, req.TempDir, webGrokMockUUID) {
		t.Fatalf("updates.jsonl missing under grok home %s for uuid %s; argv probe:\n%s", req.GrokHome, webGrokMockUUID, resp.ArgvProbe)
	}
	if resp.HTTPStatus != 200 {
		t.Fatalf("session detail HTTP %d body=%q", resp.HTTPStatus, resp.HTTPBody)
	}
}
```