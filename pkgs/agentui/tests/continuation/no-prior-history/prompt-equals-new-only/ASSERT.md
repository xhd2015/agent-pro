## Expected

- Built prompt equals `hi` (trimmed).
- Built prompt does not contain `Previous conversation`.

## Errors

- None from `Run`.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.BuiltPrompt != "hi" {
		t.Fatalf("expected built prompt %q, got %q", "hi", resp.BuiltPrompt)
	}
	if strings.Contains(strings.ToLower(resp.BuiltPrompt), "previous conversation") {
		t.Fatalf("unexpected history prefix in first message prompt: %q", resp.BuiltPrompt)
	}
}
```