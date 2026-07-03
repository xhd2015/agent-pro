## Expected

- Non-zero exit code.
- Stderr mentions `grok` (not found or similar).
- Stdout does not contain both success usage lines.

## Side Effects

- None.

## Errors

- Grok binary resolution failure.

## Exit Code

- Non-zero.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertError(t, resp, "grok")
	assertStdoutNotSuccessLines(t, resp)
}
```