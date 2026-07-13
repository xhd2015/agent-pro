## Expected

- Response is **not** the SPA static success document:
  - Must **not** be HTTP 200 HTML with `#root` as the primary SPA shell behavior for this API path when unauthenticated, **or**
  - Prefer: status is **401** (explicit token) and body is not an HTML SPA shell with bootstrap.
- Specifically: body must not look like successful static `index.html` delivery for `/` (no treating API as SPA fallback success).
  - Fail if status==200 **and** `htmlHasRootMount` **and** content-type is HTML (SPA mis-route).
  - Pass if status is 401/404/JSON 200 without SPA shell markers.

## Side Effects

- Background web cleaned up.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	// Explicit token + no Bearer: health must be API 401, never SPA HTML shell.
	if resp.HTTPStatus != 401 {
		t.Fatalf("GET /api/agent-run/health without Bearer: expected 401, got %d ct=%q body=%q",
			resp.HTTPStatus, resp.HTTPContentType, resp.HTTPBody)
	}
	if htmlHasRootMount(resp.HTTPBody) && strings.Contains(strings.ToLower(resp.HTTPContentType), "text/html") {
		t.Fatalf("API 401 body must not be SPA HTML shell: %q", resp.HTTPBody)
	}
	if htmlHasSessionBootstrap(resp.HTTPBody) {
		t.Fatalf("API path must not inject session bootstrap: body=%q", resp.HTTPBody)
	}
}
```
