---
label: unit
explanation: session reply via SLACK_MSG_SESSION_ID + SLACK_MSG_CONFIG env
---

## Expected

- Exit code 0.
- OK line for channel; trailing `\n`.
- One PostMessage: correct channel, empty ThreadTS, body present.

## Exit Code

0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	if resp.Stderr != "" {
		t.Fatalf("expected empty stderr, got:\n%s", resp.Stderr)
	}
	assert.Output(t, resp.Stdout, `---
version: 3
__TS__: type=number, example=1783398010.628649, message timestamp
---
OK ts=__TS__ channel=C0ALE44K5J6
`)
	if len(resp.PostMessages) != 1 {
		t.Fatalf("want 1 PostMessage, got %d: %+v", len(resp.PostMessages), resp.PostMessages)
	}
	post := resp.PostMessages[0]
	if post.Channel != slackTestChannelID {
		t.Fatalf("post channel = %q, want %q", post.Channel, slackTestChannelID)
	}
	if post.ThreadTS != "" {
		t.Fatalf("session reply must not set thread_ts, got %q", post.ThreadTS)
	}
	if !strings.Contains(post.Text, "env reply body") {
		t.Fatalf("post text missing body: %q", post.Text)
	}
}
```
