## Expected

- Both invocations exit non-zero.
- Stderr mentions session id and message requirement.

## Expected Output

```text
send: requires <session-id> and <message>
```

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
		t.Fatalf("send with no args should fail, got exit 0")
	}
	if resp.AltExitCode == 0 {
		t.Fatalf("send with session only should fail, got exit 0")
	}
	for _, stderr := range []string{resp.Stderr, resp.AltStderr} {
		trim := strings.TrimSpace(stderr)
		if !strings.Contains(trim, "send:") {
			t.Fatalf("stderr should mention send usage, got %q", stderr)
		}
		lower := strings.ToLower(trim)
		if !strings.Contains(lower, "session") || !strings.Contains(lower, "message") {
			t.Fatalf("stderr should mention session and message, got %q", stderr)
		}
	}
}
```