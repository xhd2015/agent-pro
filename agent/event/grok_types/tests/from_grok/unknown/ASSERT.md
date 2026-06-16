## Expected
- Empty JSON array — unknown grok event type produces no AgentEvents.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if resp.Output != `[]` && resp.Output != "null" && resp.Output != "" {
		t.Fatalf("expected empty result for unknown type, got: %s", resp.Output)
	}
}
```
