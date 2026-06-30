## Expected

- Detached `run` exits non-zero (client interrupted).
- No session id printed on the PTY capture.
- At least one session remains `running` in daemon list.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit after detach signal, got 0 combined %q", resp.Combined)
	}
	for _, line := range strings.Split(strings.TrimSpace(resp.Combined), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "session-") {
			t.Fatalf("session id should not print after detach, got %q", resp.Combined)
		}
	}
	if !resp.SessionStillRunning {
		t.Fatalf("expected running session after detach, list body %q", resp.HTTPBody)
	}
	if resp.DetachedSessionID == "" {
		t.Fatal("expected detached session id from list")
	}
}
```