## Expected

- Built prompt contains prior user text `hi`.
- Built prompt contains follow-up `what did I ask?`.

## Errors

- None from `Run`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertContains(t, resp.BuiltPrompt, "hi")
	assertContains(t, resp.BuiltPrompt, "what did I ask?")
}
```