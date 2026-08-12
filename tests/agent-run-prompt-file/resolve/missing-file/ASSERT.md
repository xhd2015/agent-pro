## Expected

- No harness error.
- API error non-empty (path missing / unreadable).
- Prompt is not treated as a successful load of body text.

## Side Effects

- Case dir may exist; target file must not.

## Errors

- Expected API error (missing/unreadable).

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertAPIError(t, resp)
	if resp.Prompt == fixturePromptBody {
		t.Fatalf("missing file must not yield fixture body; prompt=%q err=%q",
			resp.Prompt, resp.ErrString)
	}
}
```
