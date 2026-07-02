## Expected

- HTTP 200.
- JSON reports `available:true`.
- JSON keeps the route `session_id` as the web chat id.
- JSON reports `terminal_session_id:"session-1"`.

## Side Effects

- No registry file is required for the web chat id or provider runner id.

## Errors

- None from `Run`.

## Exit Code

- Test process exits non-zero until terminal resolution uses the stored mapping.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertHTTPStatus(t, resp, 200)
	requireTerminalMappingAvailable(t, req, resp.HTTPBody)
	if _, err := os.Stat(filepath.Join(req.Home, req.Runner+"-registry", req.ChatSessionID+".json")); !os.IsNotExist(err) {
		t.Fatalf("test must not rely on registry named by chat id: %v", err)
	}
	if _, err := os.Stat(filepath.Join(req.Home, req.Runner+"-registry", req.RunnerSessionID+".json")); !os.IsNotExist(err) {
		t.Fatalf("test must not rely on registry named by runner_session_id: %v", err)
	}
}
```
