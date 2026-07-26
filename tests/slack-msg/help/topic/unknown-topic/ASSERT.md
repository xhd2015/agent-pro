## Expected

- Exit code 1.
- Stdout empty.
- Stderr contains `unknown help topic:` and the topic name `not-a-topic`.

## Exit Code

1

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 1)
	if resp.Stdout != "" {
		t.Fatalf("expected empty stdout, got:\n%s", resp.Stdout)
	}
	low := strings.ToLower(resp.Stderr)
	if !strings.Contains(low, "unknown help topic") && !strings.Contains(low, "unknown topic") {
		t.Fatalf("stderr should report unknown help topic, got:\n%s", resp.Stderr)
	}
	if !strings.Contains(resp.Stderr, "not-a-topic") {
		t.Fatalf("stderr should include topic name not-a-topic, got:\n%s", resp.Stderr)
	}
}
```
