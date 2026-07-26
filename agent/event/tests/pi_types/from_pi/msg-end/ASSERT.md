## Expected
- Single AgentEvent with ActionMessage and PhaseEnd.
- **Text field is empty (no Delta available), NOT the full Content text "Hello".**
- For message_end, deltas are already shown via message_update; full text should not be repeated.

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
	assertContains(t, resp.Output, `"phase":"end"`)
	// After fix: Text is empty (no delta, avoid duplicating full text)
	assertNotContains(t, resp.Output, `"text":"Hello"`)
}
```
