## Expected

- Exit code 1.
- Error indicates prompt is required (mirrors run policy without `--open`).

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
		"prompt is required",
		"prompt required",
		"requires a prompt",
		"followup",
	)
}
```
