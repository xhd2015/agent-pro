## Expected

- `RecentMessages` has exactly 3 entries.
- Messages are the last three chronologically: `msg-three`, `msg-four`, `msg-five`.
- `FormatInfoText` includes a Recent messages section with `msg-five`.

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if resp.Info == nil {
		t.Fatal("info is nil")
	}
	msgs := resp.Info.RecentMessages
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
	assertContains(t, resp.Output, "Recent messages:")
	assertContains(t, resp.Output, "msg-five")
}
```