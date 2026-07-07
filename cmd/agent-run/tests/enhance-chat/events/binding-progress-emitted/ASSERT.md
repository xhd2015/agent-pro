## Expected

- `events.jsonl` contains a `think` event with text exactly `Resolve session id...`.
- Session reaches `finished` status (successful bind completes).

## Side Effects

- `events.jsonl` also contains user message and protocol-derived assistant content from mock `updates.jsonl`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !eventsHaveThinkText(resp.EventsParsed, resolveThinkText) {
		t.Fatalf("events.jsonl missing think %q; events=%v", resolveThinkText, resp.EventsParsed)
	}
	if req.BindingOutcome != "success" {
		t.Fatalf("expected success binding outcome, got %q", req.BindingOutcome)
	}
}
```