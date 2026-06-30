## Expected

- WaitSession returns nil after the remote `sleep 2` command exits.
- No panic and no error mentioning repeated websocket reads.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("WaitSession failed: stderr %q", resp.Stderr)
	}
	if strings.Contains(resp.Stderr, "repeated read on failed websocket connection") {
		t.Fatalf("WaitSession hit websocket read retry panic: %q", resp.Stderr)
	}
	if resp.Stderr != "" {
		t.Fatalf("expected nil error from WaitSession, got %q", resp.Stderr)
	}
	if !strings.HasPrefix(resp.Stdout, "session-") {
		t.Fatalf("expected session id in stdout, got %q", resp.Stdout)
	}
}
```