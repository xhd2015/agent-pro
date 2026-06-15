## Expected
- Health check succeeds after `EnsureServer`: `health_status` is 200, `health_err` is empty.
- The server PID is non-zero (a real process was started).

## Exit Code
- 0

```go
import (
	"encoding/json"
	"testing"
)

type lifecycleResult struct {
	HealthStatus int    `json:"health_status"`
	HealthErr    string `json:"health_err"`
	ServerPID    int    `json:"server_pid"`
}

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("server-lifecycle failed: %v", err)
	}
	if resp.Output == "" {
		t.Fatal("expected non-empty Output")
	}
	var r lifecycleResult
	if err := json.Unmarshal([]byte(resp.Output), &r); err != nil {
		t.Fatalf("failed to parse lifecycle result: %v\noutput: %s", err, resp.Output)
	}
	if r.HealthStatus != 200 {
		t.Fatalf("expected health_status 200, got %d", r.HealthStatus)
	}
	if r.HealthErr != "" {
		t.Fatalf("expected no health error, got: %s", r.HealthErr)
	}
	if r.ServerPID <= 0 {
		t.Fatalf("expected server_pid > 0, got %d", r.ServerPID)
	}
}
```
