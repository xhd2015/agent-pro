## Expected

- No API error.
- Follow-up line does **not** contain `--color` as a token.
- Open profile still present.

## Side Effects

- None (pure).

## Errors

- None.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	fu := resp.FollowUp
	if hasColorToken(fu) {
		t.Fatalf("FollowUp line must not contain --color when Color=false; got %q", fu)
	}
	assertContains(t, fu, "sess-color-false")
	assertContains(t, fu, "--open")
}
```
