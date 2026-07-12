## Expected

- Exit code 1.
- Stderr (or stdout) mentions that `--auto-send-or-resume` requires `--session-id`
  (or `--session`).

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
		"requires --session-id",
		"require --session-id",
		"requires --session",
		"--session-id",
	)
	// Must still be about auto / session-id, not an unrelated parse crash only.
	assertContainsAny(t, combined,
		"auto-send-or-resume",
		"session-id",
		"session",
	)
}
```
