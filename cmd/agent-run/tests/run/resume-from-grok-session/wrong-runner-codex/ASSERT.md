## Expected

- Exit code 1.
- Error indicates resume-from-grok-session requires `grok-tty` (codex-tty rejected).

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
		"requires grok-tty",
		"require grok-tty",
		"only grok-tty",
		"must be grok-tty",
		"grok-tty",
	)
	// Should not look like a pure "not found" of the Grok UUID once green.
	// (While RED / unknown flag, this assertion still fails as intended.)
}
```
