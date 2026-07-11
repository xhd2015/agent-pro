## Expected

- Exit code 0.
- Stdout documents `--dir` (workspace / working directory wording optional).
- Stdout ends with a trailing newline `\n`.

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
	assertContains(t, resp.Stdout, "--dir")
	if !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("help stdout must end with trailing newline; last bytes %q", tailHelp(resp.Stdout, 8))
	}
}

func tailHelp(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}
```
