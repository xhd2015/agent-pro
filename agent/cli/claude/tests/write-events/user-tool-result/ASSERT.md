## Expected
- Zero AgentEvents are emitted (`resp.Lines` is empty).

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Lines) != 0 {
		t.Fatalf("expected 0 AgentEvents for a user tool_result line, got %d: %v", len(resp.Lines), resp.Lines)
	}
}
```
