---
label: unit
explanation: --json document with messages oldest→newest and has_more
---

## Expected Output

```json
{"messages":[{"ts":"1710000001.000100","user":"U_OLDER","text":"first message"},{"ts":"1710000002.000200","user":"U_NEWER","text":"second message"},{"ts":"1710000003.000300","user":"U_NEWEST","text":"third message"}],"has_more":false}
```

## Expected

- Exit code 0.
- Single JSON document on stdout (trailing newline).
- Messages oldest→newest; `has_more` false.
- Stderr empty.

## Exit Code

0

```go
import (
	"encoding/json"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	if resp.Stderr != "" {
		t.Fatalf("expected empty stderr, got:\n%s", resp.Stderr)
	}
	var doc struct {
		Messages []struct {
			TS       string `json:"ts"`
			User     string `json:"user"`
			Text     string `json:"text"`
			ThreadTS string `json:"thread_ts"`
		} `json:"messages"`
		HasMore bool `json:"has_more"`
	}
	if err := json.Unmarshal([]byte(resp.Stdout), &doc); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, resp.Stdout)
	}
	if doc.HasMore {
		t.Fatalf("has_more = true, want false")
	}
	if len(doc.Messages) != 3 {
		t.Fatalf("messages len = %d, want 3: %+v", len(doc.Messages), doc.Messages)
	}
	want := []struct{ ts, user, text string }{
		{"1710000001.000100", "U_OLDER", "first message"},
		{"1710000002.000200", "U_NEWER", "second message"},
		{"1710000003.000300", "U_NEWEST", "third message"},
	}
	for i, w := range want {
		m := doc.Messages[i]
		if m.TS != w.ts || m.User != w.user || m.Text != w.text {
			t.Fatalf("messages[%d] = {%q,%q,%q}, want {%q,%q,%q}", i, m.TS, m.User, m.Text, w.ts, w.user, w.text)
		}
	}
	if resp.Stdout == "" || resp.Stdout[len(resp.Stdout)-1] != '\n' {
		t.Fatalf("stdout must end with trailing newline, got %q", resp.Stdout)
	}
}
```
