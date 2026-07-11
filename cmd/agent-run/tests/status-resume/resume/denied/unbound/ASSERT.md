## Expected

- Exit code 1.
- Error indicates runner session not bound / missing runner_session_id.

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
		"not bound",
		"unbound",
		"runner_session_id",
		"no runner session",
		"missing runner",
		"not resolved",
	)
}
```
