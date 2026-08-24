## Expected

- stderr starts with `warning:` and mentions ambiguous; Contents used; stdout is iTerm text.

```go
import (
	"strings"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	if len(resp.AgentRunCalls) != 1 {
		t.Fatalf("AgentRunCalls = %v, want one", resp.AgentRunCalls)
	}
	if len(resp.ContentsCalls) != 1 {
		t.Fatalf("ContentsCalls = %v, want one", resp.ContentsCalls)
	}
	if !strings.HasPrefix(strings.TrimSpace(resp.Stderr), "warning:") {
		t.Fatalf("stderr missing warning: prefix:\n%s", resp.Stderr)
	}
	if !strings.Contains(resp.Stderr, "ambiguous") {
		t.Fatalf("stderr missing ambiguous:\n%s", resp.Stderr)
	}
	assert.Output(t, resp.Stdout, `---
version: 3
---
after ambiguous warn
`)
}
```