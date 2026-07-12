## Expected

- Exit code 1.
- Stderr indicates an unknown / unrecognized command or subcommand.

## Exit Code

1

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 1)
	combined := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	if !strings.Contains(combined, "unknown") &&
		!strings.Contains(combined, "unrecognized") &&
		!strings.Contains(combined, "not-a-real-pty-cmd") {
		t.Fatalf("expected unknown-subcommand error mentioning the bad name; stderr:\n%s\nstdout:\n%s",
			resp.Stderr, resp.Stdout)
	}
}
```
