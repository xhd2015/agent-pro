---
label: unit
explanation: --reply-prefix prepended to PostMessage text
---

## Expected

- PostMessage text starts with `[bot]`.
- Also contains mock agent reply body.

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
	text := resp.PostMessages[0].Text
	if !strings.HasPrefix(text, "[bot]") {
		t.Fatalf("reply should start with prefix, got %q", text)
	}
	if !strings.Contains(text, defaultAgentReply) {
		t.Fatalf("reply missing agent body %q: %q", defaultAgentReply, text)
	}
}
```