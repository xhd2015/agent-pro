## Expected

- PUT **200**.
- POST sessions **200** or **202**.
- Response `session.workspace` equals `SelectPath` (not merely process cwd unless identical).

## Errors

- Pre-impl: PUT fails or create still uses process cwd (RED).

```go
import (
	"encoding/json"
	"net/http"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
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
	create, ok := findHTTPResult(resp, "create")
	if !ok {
		t.Fatal("missing create result")
	}
	if create.Status != http.StatusAccepted && create.Status != http.StatusOK {
		t.Fatalf("POST sessions expected 200/202, got %d body=%q", create.Status, truncate(create.Body, 400))
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(create.Body), &parsed); err != nil {
		t.Fatalf("parse create body: %v body=%q", err, truncate(create.Body, 300))
	}
	sess, _ := parsed["session"].(map[string]any)
	if sess == nil {
		t.Fatalf("create response missing session object: %q", truncate(create.Body, 300))
	}
	ws, _ := sess["workspace"].(string)
	if !pathsEqual(ws, req.SelectPath) {
		t.Fatalf("session.workspace: got %q want selected %q", ws, req.SelectPath)
	}
}
```
