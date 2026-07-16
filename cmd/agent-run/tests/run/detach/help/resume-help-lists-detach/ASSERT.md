## Expected

- Exit code 0.
- Stdout documents `--detach`.
- Stdout ends with trailing newline `\n`.

## Exit Code

0

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
		t.Fatalf("exit code = %d, stderr:\n%s", resp.ExitCode, resp.Stderr)
	}
	assertContains(t, resp.Stdout, "--detach")
	if !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("resume help stdout must end with trailing newline; last bytes %q", tailBytesResume(resp.Stdout, 8))
	}
}

func tailBytesResume(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}
```
