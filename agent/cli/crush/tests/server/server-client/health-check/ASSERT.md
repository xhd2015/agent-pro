## Expected
- `ensureServer` returns without error.
- Health endpoint (`/v1/health`) returns HTTP 200.
- Server process is alive.

## Exit Code
- 0

```go
import (
	"encoding/json"
	"testing"
)

type healthResult struct {
	Status int `json:"status"`
}

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("health-check failed: %v", err)
	}
	if resp.Output == "" {
		t.Fatal("expected non-empty Output with health result")
	}
	var result healthResult
	if err := json.Unmarshal([]byte(resp.Output), &result); err != nil {
		t.Fatalf("failed to parse health result: %v\noutput: %s", err, resp.Output)
	}
	if result.Status != 200 {
		t.Fatalf("expected health status 200, got %d", result.Status)
	}
}
```
