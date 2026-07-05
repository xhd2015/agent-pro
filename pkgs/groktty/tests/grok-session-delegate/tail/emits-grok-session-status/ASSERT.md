## Expected

- First tool event: `extensions.grok_session.status=pending`.
- Second tool event: `extensions.grok_session.status=completed`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	tools := toolEvents(resp.Events)
	if len(tools) != 2 {
		t.Fatalf("tool event count: got %d want 2\n%s", len(tools), formatEvents(resp.Events))
	}
	if got := grokStatus(tools[0]); got != "pending" {
		t.Fatalf("first tool status: got %q want pending (%#v)", got, tools[0])
	}
	if got := grokStatus(tools[1]); got != "completed" {
		t.Fatalf("second tool status: got %q want completed (%#v)", got, tools[1])
	}
}
```