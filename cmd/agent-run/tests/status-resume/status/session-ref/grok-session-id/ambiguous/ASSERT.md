## Expected

- Exit code 1.
- Combined output indicates ambiguity (ambiguous / multiple / more than one).
- Both agent-run session ids (`test-gsid-amb-a`, `test-gsid-amb-b`) are mentioned.

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
	assertContainsAny(t, combined,
		"ambiguous",
		"multiple",
		"more than one",
		"matches",
	)
	assertContains(t, combined, "test-gsid-amb-a")
	assertContains(t, combined, "test-gsid-amb-b")
}
```
