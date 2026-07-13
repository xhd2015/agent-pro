## Expected

- HTTP **200**.
- `sessions` is an array of length **5** (all seeded).
- First element `session_id` is **`sess-epsilon`** (newest `updated_at`).
- Second is **`sess-delta`** (next newest).
- If `has_more` is present, it is **false**.
- If `total` is present, it equals **5**.

## Errors

- Pre-impl: order not newest-first (RED) even if length is 5.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	r := requireOK200(t, resp, "list")
	sessions := sessionsFromBody(t, r.Body)
	if len(sessions) != 5 {
		t.Fatalf("sessions length: got %d want 5 ids=%v", len(sessions), sessionIDs(sessions))
	}
	ids := sessionIDs(sessions)
	if ids[0] != "sess-epsilon" {
		t.Fatalf("first session (newest): got %q want sess-epsilon; order=%v", ids[0], ids)
	}
	if ids[1] != "sess-delta" {
		t.Fatalf("second session: got %q want sess-delta; order=%v", ids[1], ids)
	}
	m := parseJSONMap(t, r.Body)
	if hasMore, ok := jsonBool(m, "has_more"); ok && hasMore {
		t.Fatalf("has_more should be false when limit omitted, got true")
	}
	if total, ok := jsonFloat(m, "total"); ok && int(total) != 5 {
		t.Fatalf("total: got %v want 5", total)
	}
}
```
