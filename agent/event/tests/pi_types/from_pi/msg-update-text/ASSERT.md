## Expected
- Single AgentEvent with ActionMessage type and PhaseUpdate.
- **Text field uses Delta (" world"), NOT the full Content text ("Hello").**

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, resp.Output, `"type":"message"`)
	assertContains(t, resp.Output, `"phase":"update"`)
	// After fix: Text = Delta " world", not Content[0].Text "Hello"
	assertContains(t, resp.Output, `"text":" world"`)
	assertNotContains(t, resp.Output, `"text":"Hello"`)
}
```
