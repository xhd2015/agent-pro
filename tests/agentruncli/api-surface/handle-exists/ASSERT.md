## Expected

- Package `agentruncli` imports and `Handle` is callable (`HandleCalled`).
- No harness error (package must exist and export `Handle`).
- Help path may return nil Handle error; non-nil only if implementer diverges
  (this leaf only requires callability — help content is asserted under
  `handle/help-lists-commands`).

## Side Effects

- None beyond temporary stdout/stderr capture around Handle.

## Errors

- No harness error.

## Exit Code

N/A (package call)

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp == nil {
		t.Fatal("nil response")
	}
	if !resp.HandleCalled {
		t.Fatal("expected Handle to be invoked")
	}
}
```
