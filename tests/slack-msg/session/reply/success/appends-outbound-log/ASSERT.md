---
label: unit
explanation: session reply appends direction=out to messages.jsonl
---

## Expected

- Exit code 0.
- OK line present.
- messages.jsonl has prior inbound + new outbound line with text `outbound reply text`
  and `direction=out`.

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
	assertExitCode(t, resp, 0)
	if !strings.Contains(resp.Stdout, "OK ts=") {
		t.Fatalf("stdout missing OK line:\n%s", resp.Stdout)
	}
	msgs, readErr := readMessagesJSONL(t, req.HomeDir, sessionReplyFixtureID)
	if readErr != nil {
		t.Fatalf("read messages.jsonl: %v", readErr)
	}
	if len(msgs) < 2 {
		t.Fatalf("want >=2 log lines after reply, got %d: %+v", len(msgs), msgs)
	}
	last := msgs[len(msgs)-1]
	if last.Direction != "out" {
		t.Fatalf("last direction = %q, want out; msgs=%+v", last.Direction, msgs)
	}
	if !strings.Contains(last.Text, "outbound reply text") {
		t.Fatalf("last text missing outbound body: %+v", last)
	}
}
```
