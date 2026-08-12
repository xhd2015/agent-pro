## Expected

- No harness error; no API error.
- Prompt equals `""` (empty string after TrimSpace).

## Side Effects

- Case-local empty file under `d.DOCTEST_CASE` only.

## Errors

- None (empty file is success with empty prompt — same as empty positional for
  later open/detach rules; those rules are out of this leaf).

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	assertEqual(t, "Prompt", resp.Prompt, "")
}
```
