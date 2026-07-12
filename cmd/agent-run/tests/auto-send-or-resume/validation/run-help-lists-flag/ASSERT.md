## Expected

- Exit code 0.
- Stdout contains `--auto-send-or-resume`.
- Stdout ends with trailing newline `\n`.

## Exit Code

0

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout, "--auto-send-or-resume")
	assertTrailingNewline(t, resp.Stdout, "run help stdout")
}
```
