---
label: unit
explanation: PostMessage includes thread_ts from inbound event
---

## Expected

- At least one PostMessage captured.
- `thread_ts` equals inbound message ts `1710000400.000100`.
- Reply text contains mock agent output.

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
	if post.ThreadTS != "1710000400.000100" {
		t.Fatalf("thread_ts = %q, want %q", post.ThreadTS, "1710000400.000100")
	}
	if post.Channel != slackTestChannelID {
		t.Fatalf("channel = %q, want %q", post.Channel, slackTestChannelID)
	}
	if !strings.Contains(post.Text, defaultAgentReply) {
		t.Fatalf("reply text missing agent output %q: %q", defaultAgentReply, post.Text)
	}
}
```
