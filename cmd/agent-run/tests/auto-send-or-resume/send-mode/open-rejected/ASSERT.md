## Expected

- Exit code 1.
- Error indicates `--open` is not valid for the live auto path (use run/resume only,
  or session is live / use send).

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
		"--open",
		"open",
	)
	assertContainsAny(t, combined,
		"not valid",
		"not allowed",
		"cannot",
		"live",
		"still active",
		"use send",
		"send",
		"resume",
	)
	// Must not succeed as a silent open attach.
	if resp.ExitCode == 0 {
		t.Fatal("live auto + --open must fail")
	}
}
```
