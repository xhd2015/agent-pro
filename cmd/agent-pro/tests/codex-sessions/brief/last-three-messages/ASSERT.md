## Expected

- `RecentMessages` has exactly 3 entries.
- Messages are the last three chronologically: `msg-three`, `msg-four`, `msg-five`.
- Each entry has non-empty `Kind`, `Text`, and `Formatted` fields.

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if resp.Brief == nil {
		t.Fatal("brief is nil")
	}
	msgs := resp.Brief.RecentMessages
	if len(msgs) != 3 {
		t.Fatalf("len(recent_messages) = %d, want 3", len(msgs))
	}
	wantText := []string{"msg-three", "msg-four", "msg-five"}
	for i, want := range wantText {
		if msgs[i].Text != want {
			t.Fatalf("recent_messages[%d].Text = %q, want %q", i, msgs[i].Text, want)
		}
		if msgs[i].Kind == "" || msgs[i].Formatted == "" {
			t.Fatalf("recent_messages[%d] missing kind/formatted: %+v", i, msgs[i])
		}
	}
}
```