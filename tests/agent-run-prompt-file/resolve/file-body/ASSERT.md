## Expected

- No harness error; no API error.
- Prompt equals `hello` (`fixturePromptBody`) — TrimSpace of file body, not raw
  `"  hello\n"`.

## Side Effects

- Case-local file under `d.DOCTEST_CASE` only.

## Errors

- None.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	assertEqual(t, "Prompt", resp.Prompt, fixturePromptBody)
}
```
