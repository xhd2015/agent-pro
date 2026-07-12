## Expected

- Exit code 0.
- JSON `sessions` array length 10.
- Sorted newest first: first item `session_id=sess_14`.
- Each item includes at least `runner`, `session_id`, `status`.

```go
import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("expected trailing newline, got %q", resp.Stdout)
	}
	var payload struct {
		Sessions []map[string]any `json:"sessions"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(resp.Stdout)), &payload); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, resp.Stdout)
	}
	if len(payload.Sessions) != 10 {
		t.Fatalf("expected 10 sessions, got %d", len(payload.Sessions))
	}
	for i := 0; i < 10; i++ {
		wantID := fmt.Sprintf("sess_%02d", 14-i)
		gotID, _ := payload.Sessions[i]["session_id"].(string)
		if gotID != wantID {
			t.Fatalf("sessions[%d].session_id=%q want %q", i, gotID, wantID)
		}
		if _, ok := payload.Sessions[i]["runner"]; !ok {
			t.Fatalf("sessions[%d] missing runner", i)
		}
		if _, ok := payload.Sessions[i]["status"]; !ok {
			t.Fatalf("sessions[%d] missing status", i)
		}
		if strings.Contains(gotID, "/") {
			t.Fatalf("compound session_id %q", gotID)
		}
	}
}
```
