## Expected

- Exit code 0.
- Stdout mentions the spawned serve PID and/or session id.
- Serve process is still alive after dry-run (not killed).
- Stdout ends with trailing newline `\n`.

## Side Effects

- Spawned test serve remains alive until harness cleanup.

## Exit Code

0

```go
import (
	"strconv"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	if resp.ServePID <= 0 {
		t.Fatalf("expected spawned serve PID in response")
	}
	out := resp.Stdout
	pidStr := strconv.Itoa(resp.ServePID)
	hasPID := strings.Contains(out, pidStr)
	hasSession := req.SpawnedSessionID != "" && strings.Contains(out, req.SpawnedSessionID)
	if !hasPID && !hasSession {
		t.Fatalf("dry-run stdout must list spawned PID %s and/or session %q; got:\n%s",
			pidStr, req.SpawnedSessionID, out)
	}
	if !resp.ServeAliveAfter {
		t.Fatalf("dry-run must not kill serve pid %d", resp.ServePID)
	}
	assertTrailingNewline(t, resp.Stdout, "kill-orphans --dry-run stdout")
}
```
