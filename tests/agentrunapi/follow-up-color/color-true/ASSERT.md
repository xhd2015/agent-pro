## Expected

- No API error.
- Follow-up line contains token `--color`.
- Session id and open profile still present; no `--new-terminal`.

## Side Effects

- None (pure).

## Errors

- None.

## Exit Code

N/A

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	fu := resp.FollowUp
	if !hasColorToken(fu) {
		t.Fatalf("FollowUp line must contain --color token; got %q", fu)
	}
	assertContains(t, fu, "sess-color-true")
	assertContains(t, fu, "--open")
	if strings.Contains(fu, "--new-terminal") {
		t.Fatalf("must not emit --new-terminal; got %q", fu)
	}
}
```
