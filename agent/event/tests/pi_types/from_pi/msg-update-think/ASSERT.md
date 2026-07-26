## Expected
- Single AgentEvent with ActionThink and PhaseUpdate.
- **Text field uses Delta (" deeper"), NOT the full Content[0].Thinking ("think").**

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, resp.Output, `"type":"think"`)
	assertContains(t, resp.Output, `"phase":"update"`)
	// After fix: Text = Delta " deeper", not Content[0].Thinking "think"
	assertContains(t, resp.Output, `"text":" deeper"`)
	assertNotContains(t, resp.Output, `"text":"think"`)
}
```
