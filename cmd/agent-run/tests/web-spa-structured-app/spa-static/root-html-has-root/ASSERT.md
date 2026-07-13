## Expected

- HTTP status **200**.
- Body is HTML SPA shell: contains mount node `id="root"` (or `id='root'`).
- Content-Type is HTML when set (`text/html`).

## Side Effects

- Background `agent-run web` stopped on cleanup.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if resp.HTTPStatus != 200 {
		t.Fatalf("GET / expected 200, got %d body=%q", resp.HTTPStatus, truncate(resp.HTTPBody, 200))
	}
	if !htmlHasRootMount(resp.HTTPBody) {
		t.Fatalf("GET / body missing #root mount, content-type=%q body=%q",
			resp.HTTPContentType, truncate(resp.HTTPBody, 400))
	}
	if ct := strings.ToLower(resp.HTTPContentType); ct != "" && !strings.Contains(ct, "text/html") {
		t.Fatalf("expected text/html content-type, got %q", resp.HTTPContentType)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
```
