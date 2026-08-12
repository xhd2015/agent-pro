## Expected

- No harness error.
- API error non-empty (mutually exclusive).
- Error message mentions `prompt-file` and exclusive intent (`exclusive` or
  `mutually`).
- Prompt left empty or unused (do not require a successful body load).

## Side Effects

- Case-local file under `d.DOCTEST_CASE` only.

## Errors

- Expected API error (exclusive).

## Exit Code

N/A

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertAPIError(t, resp)
	msg := strings.ToLower(resp.ErrString)
	if !strings.Contains(msg, "prompt-file") {
		t.Fatalf("exclusive error should mention prompt-file; got %q", resp.ErrString)
	}
	if !strings.Contains(msg, "exclusive") && !strings.Contains(msg, "mutually") {
		t.Fatalf("exclusive error should mention exclusive/mutually; got %q", resp.ErrString)
	}
}
```
