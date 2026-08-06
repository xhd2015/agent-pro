## Expected

- Non-nil error indicating ambiguous match / need to specify runner.
- Error mentions runner (or "ambiguous" / multiple matches).
- View is nil.

## Errors

- ask for runner / ambiguous

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertError(t, resp)
	msg := strings.ToLower(resp.Err.Error())
	// Accept common phrasings: runner required, ambiguous, multiple matches.
	ok := strings.Contains(msg, "runner") ||
		strings.Contains(msg, "ambiguous") ||
		strings.Contains(msg, "multiple") ||
		strings.Contains(msg, "more than one")
	if !ok {
		t.Fatalf("error %q should indicate need for runner / ambiguous match", resp.Err.Error())
	}
	if resp.View != nil {
		t.Fatalf("View should be nil on ambiguous show, got %+v", resp.View)
	}
}
```
