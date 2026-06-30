## Expected

- Exit code 0 after `sleep 3` completes.
- PTY output contains session id line.
- Output must not contain `i/o timeout` (websocket read deadline leak).

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
		t.Fatalf("exit code %d output %q", resp.ExitCode, resp.Combined)
	}
	if strings.Contains(resp.Combined, "i/o timeout") {
		t.Fatalf("attach hit websocket read timeout during quiet sleep: %q", resp.Combined)
	}
	if !strings.Contains(resp.Combined, "session-") {
		t.Fatalf("expected session id in PTY output, got %q", resp.Combined)
	}
}
```