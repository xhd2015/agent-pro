## Expected

- Built prompt contains `Previous conversation` (or equivalent current prefix; match case-insensitive).
- Built prompt contains prior user text `hi`.
- Built prompt contains new prompt `follow-up please`.

## Errors

- None from `Run` once `ResolveRunnerPrompt` exists.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.ToLower(resp.BuiltPrompt), "previous conversation") {
		t.Fatalf("unbound multi-turn expected history prefix, got:\n%s", resp.BuiltPrompt)
	}
	assertContains(t, resp.BuiltPrompt, "hi")
	assertContains(t, resp.BuiltPrompt, "follow-up please")
}
```
