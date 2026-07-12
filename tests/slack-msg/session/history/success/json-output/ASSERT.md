---
label: unit
explanation: session history --json chronological document
---

## Expected

- Exit code 0.
- JSON on stdout with messages oldest→newest; trailing `\n`.
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
			MessageID string `json:"message_id"`
			TS        string `json:"ts"`
			User      string `json:"user"`
			Text      string `json:"text"`
			Direction string `json:"direction"`
		} `json:"messages"`
	}
	if err := json.Unmarshal([]byte(resp.Stdout), &doc); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, resp.Stdout)
	}
	if len(doc.Messages) != 3 {
		t.Fatalf("messages len = %d, want 3: %+v", len(doc.Messages), doc.Messages)
	}
	wantIDs := []string{"m1", "m2", "m3"}
	wantText := []string{"first", "second", "third"}
	for i := range wantIDs {
		m := doc.Messages[i]
		if m.MessageID != wantIDs[i] && m.TS == "" {
			t.Fatalf("messages[%d] missing id/ts: %+v", i, m)
		}
		if m.MessageID != "" && m.MessageID != wantIDs[i] {
			t.Fatalf("messages[%d].message_id = %q, want %q", i, m.MessageID, wantIDs[i])
		}
		if m.Text != wantText[i] {
			t.Fatalf("messages[%d].text = %q, want %q", i, m.Text, wantText[i])
		}
	}
	if resp.Stdout == "" || resp.Stdout[len(resp.Stdout)-1] != '\n' {
		t.Fatalf("stdout must end with trailing newline, got %q", resp.Stdout)
	}
}
```
