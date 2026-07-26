---
label: e2e
---

## Expected

- Each GET **200**.
- **q-prompt** / **q-case**: single session `sess-delta`; `total=1` when present.
- **q-id**: single session `sess-beta`.
- **q-workspace**: single session `sess-delta`.
- **q-runner**: single session `sess-gamma`.

## Errors

- Pre-impl: `q` ignored → full list (RED).

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func assertSingleMatch(t *testing.T, resp *Response, name, wantID string) {
	t.Helper()
	r := requireOK200(t, resp, name)
	sessions := sessionsFromBody(t, r.Body)
	ids := sessionIDs(sessions)
	if len(sessions) != 1 || ids[0] != wantID {
		t.Fatalf("%s: got ids=%v want single %q body=%q", name, ids, wantID, truncate(r.Body, 300))
	}
	m := parseJSONMap(t, r.Body)
	if total, ok := jsonFloat(m, "total"); ok && int(total) != 1 {
		t.Fatalf("%s total: got %v want 1", name, total)
	}
}

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	assertSingleMatch(t, resp, "q-prompt", "sess-delta")
	assertSingleMatch(t, resp, "q-id", "sess-beta")
	assertSingleMatch(t, resp, "q-workspace", "sess-delta")
	assertSingleMatch(t, resp, "q-runner", "sess-gamma")
	assertSingleMatch(t, resp, "q-case", "sess-delta")
}
```
