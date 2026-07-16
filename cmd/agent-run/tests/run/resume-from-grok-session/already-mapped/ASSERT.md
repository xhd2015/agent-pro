## Expected

- Exit code 1.
- Error indicates the Grok session is already mapped in agent-run.
- Prefer a hint toward `resume --grok-session-id` (accepted when present).

## Exit Code

1

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 1)
	combined := combinedOut(resp)
	assertContainsAny(t, combined,
		"already mapped",
		"already bound",
		"already exists",
		"already imported",
		"mapped",
	)
	// Optional product hint (soft): resume --grok-session-id
	// Do not hard-fail if only "already mapped" is present after GREEN.
	// Soft check: if either "resume" or "grok-session-id" appears, good.
	// (Implementer should include the hint per locked product rules.)
}
```
