## Expected
- Output JSON array contains three events in order: `message`, `think`, `tool_call`.
- The message text is `hello`, the think text is `reasoning`, the tool call is `Bash` with `{"command":"ls"}`.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"message"`)
	assertContains(t, resp.Output, `"text":"hello"`)
	assertContains(t, resp.Output, `"type":"think"`)
	assertContains(t, resp.Output, `"reasoning"`)
	assertContains(t, resp.Output, `"type":"tool_call"`)
	assertContains(t, resp.Output, `"tool":"Bash"`)
	assertContains(t, resp.Output, `"tool_input":{"command":"ls"}`)

	// block order: message before think before tool_call
	msgAt := strings.Index(resp.Output, `"type":"message"`)
	thinkAt := strings.Index(resp.Output, `"type":"think"`)
	toolAt := strings.Index(resp.Output, `"type":"tool_call"`)
	if msgAt < 0 || thinkAt < 0 || toolAt < 0 || !(msgAt < thinkAt && thinkAt < toolAt) {
		t.Fatalf("expected message<think<tool_call order, got msg=%d think=%d tool=%d", msgAt, thinkAt, toolAt)
	}
}
```
