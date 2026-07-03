## Expected

- Non-zero exit code.
- Stderr mentions `timeout`.
- Stdout does not contain both success usage lines.

## Side Effects

- None.

## Errors

- Usage line wait exceeded `GROK_SHOW_USAGE_TIMEOUT`.

## Exit Code

- Non-zero.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertError(t, resp, "timeout")
	assertStdoutNotSuccessLines(t, resp)
}
```