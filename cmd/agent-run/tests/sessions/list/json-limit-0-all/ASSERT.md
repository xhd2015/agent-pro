## Expected

- Exit code 0.
- JSON `sessions` array length 15, newest first.

```go
import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	var payload struct {
		Sessions []map[string]any `json:"sessions"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(resp.Stdout)), &payload); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, resp.Stdout)
	}
	if len(payload.Sessions) != 15 {
		t.Fatalf("expected 15 sessions, got %d", len(payload.Sessions))
	}
	for i := 0; i < 15; i++ {
		wantID := fmt.Sprintf("sess_%02d", 14-i)
		gotID, _ := payload.Sessions[i]["session_id"].(string)
		if gotID != wantID {
			t.Fatalf("sessions[%d].session_id=%q want %q", i, gotID, wantID)
		}
	}
}
```
