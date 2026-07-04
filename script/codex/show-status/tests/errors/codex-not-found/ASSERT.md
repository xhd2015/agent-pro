## Expected

- Non-zero exit code.
- Stderr mentions `codex` (not found or similar).
- Stdout does not contain all three success status lines.

## Side Effects

- No persistent tty-watch session (resolution fails before run).

## Errors

- Codex binary resolution failure.

## Exit Code

- Non-zero.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertError(t, resp, "codex")
	assertStdoutNotSuccessLines(t, resp)
}
```