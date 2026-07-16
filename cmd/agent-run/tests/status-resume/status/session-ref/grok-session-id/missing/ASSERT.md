## Expected

- Exit code 1.
- Combined stderr/stdout indicates not found (or no match) for the grok session id.
- Message mentions the requested id and/or "grok".

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
		"not found",
		"no such",
		"unknown",
		"no match",
		"no session",
	)
	// Prefer explicit mention of the lookup key or grok wording.
	assertContainsAny(t, combined,
		strings.ToLower(req.RunnerSessionID),
		"grok-session-id",
		"grok session",
		"runner_session_id",
		"runner session",
	)
}
```
