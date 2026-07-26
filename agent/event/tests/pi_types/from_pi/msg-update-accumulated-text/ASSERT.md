## Expected
- Single AgentEvent with ActionMessage type and PhaseUpdate.
- **Text field contains only the Delta (" feature."), NOT the full accumulated Content[0].Text.**
- The full accumulated text string must NOT appear.

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
	// After fix: Text = Delta " feature." only
	assertContains(t, resp.Output, `"text":" feature."`)
	// The full accumulated text must NOT appear in the output
	assertNotContains(t, resp.Output, "The user has given me a detailed requirement for creating a macOS menu bar app. I need to design a comprehensive doctest tree for this feature.")
}
```
