---
label: e2e
---

## Expected

- Exit code 0.
- Stdout is valid JSON with empty `sessions` array (or empty array root).
- Output ends with trailing newline after last content.

```go
import (
	"encoding/json"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("expected trailing newline on stdout, got %q", resp.Stdout)
	}
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
		} else {
			t.Fatalf("expected sessions key with array, got %#v", v)
		}
	default:
		t.Fatalf("unexpected JSON type %T", raw)
	}
}
```
