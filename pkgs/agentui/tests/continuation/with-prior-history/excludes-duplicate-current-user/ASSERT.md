## Expected

- Built prompt still contains prior `hi`.
- Built prompt contains `what did I ask?` exactly once (not duplicated in history + tail).

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
	assertContains(t, resp.BuiltPrompt, "hi")
	assertContains(t, resp.BuiltPrompt, "what did I ask?")
	if strings.Count(resp.BuiltPrompt, "what did I ask?") != 1 {
		t.Fatalf("expected follow-up prompt once, got %d in:\n%s", strings.Count(resp.BuiltPrompt, "what did I ask?"), resp.BuiltPrompt)
	}
}
```