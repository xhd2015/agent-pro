## Expected

- Exit code 1.
- Error indicates only `grok-tty` is allowed when `--agent-runner` is set
  (`grok` is not acceptable for this import path).

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
}
```
