## Expected

- First arg is `run`.
- No `--session-id` / `--session-id=` / `--session` pair.
- Last arg is the prompt.

## Errors

- None.

## Exit Code

N/A

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	if len(resp.Args) < 2 {
		t.Fatalf("expected at least run + prompt, got %q", resp.Args)
	}
	assertEqual(t, "args[0]", resp.Args[0], "run")
	assertEqual(t, "last", resp.Args[len(resp.Args)-1], "stateless prompt")
	for _, a := range resp.Args {
		if strings.HasPrefix(a, "--session-id") || a == "--session" || a == "should-not-appear" {
			t.Fatalf("stateless argv must not include session: %q in %q", a, resp.Args)
		}
	}
}
```
