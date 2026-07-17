## Expected

- Handle returns a non-nil error (`ErrString` non-empty).
- Error text indicates unknown/unrecognized command (case-insensitive), and
  preferably includes the token `not-a-real-command`.

## Side Effects

- None required (no session files, no network).

## Errors

- Handle API error expected; harness error must be nil.

## Exit Code

N/A (package call; thin main exit 1 is out of leaf scope)

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp == nil {
		t.Fatal("nil response")
	}
	assertHandleError(t, resp)
	assertContainsFold(t, resp.ErrString, "unknown")
	assertContains(t, resp.ErrString, "not-a-real-command")
}
```
