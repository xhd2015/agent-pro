## Expected

- Exit code 0.
- Stdout documents `--resume-from-grok-session`.
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
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout, "--resume-from-grok-session")
	if !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("help stdout must end with trailing newline; last bytes %q", tailBytes(resp.Stdout, 8))
	}
}

func tailBytes(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}
```
