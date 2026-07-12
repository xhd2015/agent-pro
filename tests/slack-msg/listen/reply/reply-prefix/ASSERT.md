---
label: unit
explanation: stateless --reply-prefix prepended to PostMessage text with thread_ts
---

## Expected

- PostMessage text starts with `[bot]`.
- Contains mock agent reply body.
- `thread_ts` equals inbound message ts `1710000500.000100`.
- Channel is the test channel.

## Exit Code

0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.PostMessages) < 1 {
		t.Fatal("expected PostMessage reply")
	}
	post := resp.PostMessages[0]
	if !strings.HasPrefix(post.Text, "[bot]") {
		t.Fatalf("reply should start with prefix, got %q", post.Text)
	}
	if !strings.Contains(post.Text, defaultAgentReply) {
		t.Fatalf("reply missing agent body %q: %q", defaultAgentReply, post.Text)
	}
	if post.ThreadTS != "1710000500.000100" {
		t.Fatalf("thread_ts = %q, want %q", post.ThreadTS, "1710000500.000100")
	}
	if post.Channel != slackTestChannelID {
		t.Fatalf("channel = %q, want %q", post.Channel, slackTestChannelID)
	}
}
```
