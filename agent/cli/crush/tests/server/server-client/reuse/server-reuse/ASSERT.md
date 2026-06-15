## Expected
- Both clients can health check successfully: `health_A_status` and `health_B_status` are both 200, with empty errors.
- Client B reports the server as started (reused the one client A created).

## Exit Code
- 0

```go
import (
	"encoding/json"
	"testing"
)

type reuseResult struct {
	HealthAStatus int    `json:"health_A_status"`
	HealthAErr    string `json:"health_A_err"`
	HealthBStatus int    `json:"health_B_status"`
	HealthBErr    string `json:"health_B_err"`
	ServerBStarted bool  `json:"server_B_started"`
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
	if !r.ServerBStarted {
		t.Fatal("expected client B to have reused the started server")
	}
}
```
