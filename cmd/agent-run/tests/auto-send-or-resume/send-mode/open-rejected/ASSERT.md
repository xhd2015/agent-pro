## Expected

- Exit code 0 (live + `--open` no longer rejected).
- Stdout is `msg_N` (send path).
- Must not spawn provider `--resume`.

## Exit Code

0

```go
import (
	"os"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v\nstdout:\n%s\nstderr:\n%s", err, resp.Stdout, resp.Stderr)
	}
	assertSuccess(t, resp)

	first := strings.TrimSpace(strings.Split(resp.Stdout, "\n")[0])
	if !strings.HasPrefix(first, "msg_") {
		t.Fatalf("stdout first line must be msg_N, got %q\nstderr:\n%s", first, resp.Stderr)
	}
	assertTrailingNewline(t, resp.Stdout, "send stdout")

	// Soft: stderr may note that --open was ignored while live.
	_ = resp.Stderr

	if fileExists(req.ArgvProbePath) {
		probe, _ := os.ReadFile(req.ArgvProbePath)
		if strings.Contains(string(probe), "--resume") {
			t.Fatalf("live send must not spawn provider --resume; argv:\n%s", probe)
		}
	}
}
```
