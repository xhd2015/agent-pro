## Expected

- Exit code 0 (no panic from WaitSession retrying reads on a failed websocket).
- Stdout is a single line matching `session-N` pattern.
- Stderr does not contain `repeated read on failed websocket connection`.

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
		t.Fatalf("exit code %d stderr %q stdout %q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if strings.Contains(resp.Stderr, "repeated read on failed websocket connection") {
		t.Fatalf("WaitSession panicked on websocket read retry: %q", resp.Stderr)
	}
	if strings.Contains(resp.Combined, "panic:") {
		t.Fatalf("unexpected panic output: %q", resp.Combined)
	}
	line := strings.TrimSpace(resp.Stdout)
	if !strings.HasPrefix(line, "session-") {
		t.Fatalf("stdout should be session id only, got %q", resp.Stdout)
	}
}
```