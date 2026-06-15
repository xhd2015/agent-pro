## Expected
- Before any client: `before_count` is 0.
- After client A calls `EnsureServer`: `after_A_count` is at least 1.
- After client B calls `EnsureServer`: `after_B_count` equals `after_A_count` (no new process spawned, i.e. exactly 1).
- Both clients can health check successfully: `health_A_status` and `health_B_status` are both 200, with empty errors.

## Exit Code
- 0

```go
import (
	"encoding/json"
	"testing"
)

type reuseResult struct {
	BeforeCount    int    `json:"before_count"`
	AfterACount    int    `json:"after_A_count"`
	AfterBCount    int    `json:"after_B_count"`
	HealthAStatus  int    `json:"health_A_status"`
	HealthAErr     string `json:"health_A_err"`
	HealthBStatus  int    `json:"health_B_status"`
	HealthBErr     string `json:"health_B_err"`
}

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("server-reuse failed: %v", err)
	}
	if resp.Output == "" {
		t.Fatal("expected non-empty Output")
	}
	var r reuseResult
	if err := json.Unmarshal([]byte(resp.Output), &r); err != nil {
		t.Fatalf("failed to parse reuse result: %v\noutput: %s", err, resp.Output)
	}
	if r.BeforeCount != 0 {
		t.Fatalf("expected before_count 0, got %d", r.BeforeCount)
	}
	if r.AfterACount < 1 {
		t.Fatalf("expected after_A_count >= 1, got %d", r.AfterACount)
	}
	if r.AfterBCount != r.AfterACount {
		t.Fatalf("expected after_B_count (%d) to equal after_A_count (%d) — server should be reused", r.AfterBCount, r.AfterACount)
	}
	if r.HealthAStatus != 200 {
		t.Fatalf("expected client A health status 200, got %d", r.HealthAStatus)
	}
	if r.HealthAErr != "" {
		t.Fatalf("expected no health error for client A, got: %s", r.HealthAErr)
	}
	if r.HealthBStatus != 200 {
		t.Fatalf("expected client B health status 200, got %d", r.HealthBStatus)
	}
	if r.HealthBErr != "" {
		t.Fatalf("expected no health error for client B, got: %s", r.HealthBErr)
	}
}
```
