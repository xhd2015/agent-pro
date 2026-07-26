## Expected

- Built prompt equals `only`.
- Built prompt does not contain `Previous conversation`.

## Errors

- None from `Run` once `ResolveRunnerPrompt` exists.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.BuiltPrompt != "only" {
		t.Fatalf("expected built prompt %q, got %q", "only", resp.BuiltPrompt)
	}
	if strings.Contains(strings.ToLower(resp.BuiltPrompt), "previous conversation") {
		t.Fatalf("unexpected history prefix: %q", resp.BuiltPrompt)
	}
}
```
