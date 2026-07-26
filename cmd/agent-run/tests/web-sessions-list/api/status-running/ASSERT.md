---
label: e2e
---

## Expected

- HTTP **200**.
- Exactly **one** session; `session_id` = `sess-beta`; `status` = `running`.
- If `total` present, equals **1**.

## Errors

- Pre-impl: status ignored → all five returned (RED).

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	r := requireOK200(t, resp, "running")
	sessions := sessionsFromBody(t, r.Body)
	if len(sessions) != 1 {
		t.Fatalf("running filter length: got %d want 1 ids=%v", len(sessions), sessionIDs(sessions))
	}
	id, _ := sessions[0]["session_id"].(string)
	if id != "sess-beta" {
		t.Fatalf("running session_id: got %q want sess-beta", id)
	}
	st, _ := sessions[0]["status"].(string)
	if strings.ToLower(strings.TrimSpace(st)) != "running" {
		t.Fatalf("status field: got %q want running", st)
	}
	m := parseJSONMap(t, r.Body)
	if total, ok := jsonFloat(m, "total"); ok && int(total) != 1 {
		t.Fatalf("total: got %v want 1", total)
	}
}
```
