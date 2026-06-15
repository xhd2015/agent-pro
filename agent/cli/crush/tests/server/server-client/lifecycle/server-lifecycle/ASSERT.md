## Expected
- Before any client: `before_count` is 0 (no orphaned server).
- After `EnsureServer`: `after_ensure_count` is at least 1 (server process is running).
- Health check before kill: `health_status` is 200, `health_err` is empty.
- After manual kill: `after_kill_count` is 0 (process is gone).
- Health check after kill: `health_after_kill_err` is non-empty (health check fails).

## Exit Code
- 0

```go
import (
	"encoding/json"
	"testing"
)

type lifecycleResult struct {
	BeforeCount        int    `json:"before_count"`
	AfterEnsureCount   int    `json:"after_ensure_count"`
	HealthStatus       int    `json:"health_status"`
	HealthErr          string `json:"health_err"`
	AfterKillCount     int    `json:"after_kill_count"`
	HealthAfterKillErr string `json:"health_after_kill_err"`
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
	if r.BeforeCount != 0 {
		t.Fatalf("expected before_count 0 (no orphaned server), got %d", r.BeforeCount)
	}
	if r.AfterEnsureCount < 1 {
		t.Fatalf("expected after_ensure_count >= 1 (server running), got %d", r.AfterEnsureCount)
	}
	if r.HealthStatus != 200 {
		t.Fatalf("expected health_status 200, got %d", r.HealthStatus)
	}
	if r.HealthErr != "" {
		t.Fatalf("expected no health error before kill, got: %s", r.HealthErr)
	}
	if r.AfterKillCount != 0 {
		t.Fatalf("expected after_kill_count 0 (process killed), got %d", r.AfterKillCount)
	}
	if r.HealthAfterKillErr == "" {
		t.Fatal("expected health check to fail after kill, but got no error")
	}
}
```
