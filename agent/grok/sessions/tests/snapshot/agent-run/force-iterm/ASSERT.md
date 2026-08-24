## Expected

- AgentRun not probed; Contents used; `source=iterm`.

```go
import "encoding/json"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	if len(resp.AgentRunCalls) != 0 {
		t.Fatalf("AgentRunCalls = %v, want none with --iterm", resp.AgentRunCalls)
	}
	if len(resp.ContentsCalls) != 1 {
		t.Fatalf("ContentsCalls = %v, want one", resp.ContentsCalls)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(resp.Stdout), &got); err != nil {
		t.Fatalf("stdout not JSON: %v\n%s", err, resp.Stdout)
	}
	if got["source"] != "iterm" {
		t.Fatalf("source = %v, want iterm", got["source"])
	}
	if got["contents"] != "forced iterm pane" {
		t.Fatalf("contents = %v", got["contents"])
	}
}
```
