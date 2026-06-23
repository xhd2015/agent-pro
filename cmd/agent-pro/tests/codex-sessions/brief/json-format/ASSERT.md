## Expected

- JSON unmarshals with `id`, `started_at`, `cwd`, `path`, and `recent_messages`.
- `recent_messages` has at least one entry with `kind`, `text`, `formatted`.
- Last message text is `second insight`.

## Errors

- None.

```go
import (
	"encoding/json"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	var payload struct {
		ID             string `json:"id"`
		StartedAt      string `json:"started_at"`
		CWD            string `json:"cwd"`
		Path           string `json:"path"`
		RecentMessages []struct {
			Kind      string `json:"kind"`
			Text      string `json:"text"`
			Formatted string `json:"formatted"`
		} `json:"recent_messages"`
	}
	if err := json.Unmarshal(resp.JSON, &payload); err != nil {
		t.Fatalf("unmarshal brief json: %v\n%s", err, string(resp.JSON))
	}
	if payload.ID != "01900007-dddd-7ddd-dddd-dddddddddddd" {
		t.Fatalf("id = %q, want fixture uuid", payload.ID)
	}
	if payload.CWD != "/tmp/json-brief" || payload.StartedAt == "" {
		t.Fatalf("unexpected metadata: %+v", payload)
	}
	if payload.Path == "" || !strings.Contains(payload.Path, "01900007-dddd-7ddd-dddd-dddddddddddd") {
		t.Fatalf("path = %q, want rollout jsonl path containing session id", payload.Path)
	}
	if len(payload.RecentMessages) < 1 {
		t.Fatal("recent_messages is empty")
	}
	last := payload.RecentMessages[len(payload.RecentMessages)-1]
	if last.Text != "second insight" {
		t.Fatalf("last message text = %q, want second insight", last.Text)
	}
	if last.Kind == "" || last.Formatted == "" {
		t.Fatalf("recent_messages entry missing fields: %+v", last)
	}
}
```