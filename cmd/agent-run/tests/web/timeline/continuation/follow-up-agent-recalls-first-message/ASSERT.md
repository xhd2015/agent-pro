---
label: e2e
---

## Expected

- HTTP **200** on session detail.
- At least one assistant `message` event text mentions `hi` (case-insensitive), from continuation-aware run.

## Side Effects

- `events.jsonl` contains two user prompts and multiple assistant messages.

## Errors

- None from `Run`.

```go
import (
	"strings"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.HTTPStatus != 200 {
		t.Fatalf("expected HTTP 200, got %d body=%q", resp.HTTPStatus, resp.HTTPBody)
	}
	body := waitForAssistantMention(t, req, req.SessionRunner, req.SessionID, "hi", 5*time.Second)
	if !assistantMessagesMentioning(body, "hi") {
		t.Fatalf("assistant did not recall first message hi: %s", body)
	}
	if !strings.Contains(body, req.FollowUpPrompt) {
		t.Fatalf("detail missing follow-up user prompt %q: %s", req.FollowUpPrompt, body)
	}
}
```