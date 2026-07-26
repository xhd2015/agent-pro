## Expected

- Built prompt equals trimmed `follow-up please`.
- Built prompt does **not** contain `Previous conversation` (case-insensitive).
- Built prompt does **not** need to contain prior user text `hi` (and must not dump the transcript).

## Errors

- None from `Run` once `ResolveRunnerPrompt` exists (compile/link fails RED until implementer adds the API).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	want := "follow-up please"
	if resp.BuiltPrompt != want {
		t.Fatalf("expected built prompt %q, got %q", want, resp.BuiltPrompt)
	}
	if strings.Contains(strings.ToLower(resp.BuiltPrompt), "previous conversation") {
		t.Fatalf("bound resume must not inject history prefix, got: %q", resp.BuiltPrompt)
	}
	assertNotContains(t, resp.BuiltPrompt, "hi")
}
```
