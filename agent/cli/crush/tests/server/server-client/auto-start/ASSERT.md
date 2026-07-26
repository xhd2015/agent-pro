## Expected
- `ensureServer` returns without error.
- Server is now running and `/v1/health` returns 200.
- Only one server instance is started (deterministic socket).

```go
import (
	"encoding/json"
	"testing"
)

type healthResult struct {
	Status int  `json:"status"`
	Started bool `json:"started"`
}

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("auto-start failed: %v", err)
	}
	if resp.Output == "" {
		t.Fatal("expected non-empty Output")
	}
	var result healthResult
	if err := json.Unmarshal([]byte(resp.Output), &result); err != nil {
		t.Fatalf("failed to parse result: %v\noutput: %s", err, resp.Output)
	}
	if result.Status != 200 {
		t.Fatalf("expected health status 200 after auto-start, got %d", result.Status)
	}
	if !result.Started {
		t.Fatal("expected server to have been started by ensureServer")
	}
}
```
