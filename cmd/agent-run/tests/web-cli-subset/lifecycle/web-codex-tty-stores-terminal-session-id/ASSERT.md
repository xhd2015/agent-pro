## Expected

- Session detail JSON includes non-empty `terminal_session_id`.
- Registry file exists for that terminal id under `codex-tty-registry/`.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	parsed := decodeJSONBody(t, resp.HTTPBody)
	sess, _ := parsed["session"].(map[string]any)
	if sess == nil {
		t.Fatalf("missing session object: %s", resp.HTTPBody)
	}
	termID := stringField(sess, "terminal_session_id")
	if termID == "" {
		t.Fatalf("missing terminal_session_id in session detail: %s", resp.HTTPBody)
	}
	regPath := filepath.Join(req.Home, "codex-tty-registry", termID+".json")
	if _, statErr := os.Stat(regPath); statErr != nil {
		t.Fatalf("registry entry missing for terminal_session_id %q: %v", termID, statErr)
	}
}
```
