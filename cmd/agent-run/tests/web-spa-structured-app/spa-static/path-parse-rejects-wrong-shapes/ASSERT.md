## Expected

- Every wrong-shape path:
  - May return **200** SPA HTML (fallback is fine), **or** other non-error status.
  - Must **not** include `agent-run-session-bootstrap`.
- Documents G5 reject side of `parseSessionPagePath` via bootstrap absence (accept path covered by `session-path/seeded-bootstrap`).

## Side Effects

- Seeded `sessions/spa-parse-id/`; web cleanup.

```go
import (
	"fmt"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if len(resp.HTTPResults) == 0 {
		t.Fatal("expected multi-path HTTPResults")
	}
	var bad []string
	for _, r := range resp.HTTPResults {
		if htmlHasSessionBootstrap(r.Body) {
			bad = append(bad, fmt.Sprintf("%s status=%d", r.Path, r.Status))
		}
	}
	if len(bad) > 0 {
		t.Fatalf("wrong path shapes must not inject bootstrap: %v", bad)
	}
}
```
