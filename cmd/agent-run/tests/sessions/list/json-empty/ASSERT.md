## Expected

- Exit code 0.
- Stdout is valid JSON representing an empty session list (array or object with empty `sessions`).

```go
import (
	"encoding/json"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	text := strings.TrimSpace(resp.Stdout)
	var raw any
	if err := json.Unmarshal([]byte(text), &raw); err != nil {
		t.Fatalf("invalid JSON stdout: %v\n%s", err, text)
	}
	switch v := raw.(type) {
	case []any:
		if len(v) != 0 {
			t.Fatalf("expected empty array, got %d entries", len(v))
		}
	case map[string]any:
		if sessions, ok := v["sessions"].([]any); ok {
			if len(sessions) != 0 {
				t.Fatalf("expected empty sessions list, got %d", len(sessions))
			}
		}
	default:
		t.Fatalf("unexpected JSON type %T", raw)
	}
}
```