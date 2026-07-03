## Expected

- `Escape` succeeds without error.
- Escaped output does not contain raw newline characters.
- Newlines are represented with an AppleScript-safe escape (e.g. `\n` sequence).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Err != nil {
		t.Fatal(resp.Err)
	}
	assertEscapedNoRawNewline(t, resp.Escaped)
	if resp.Escaped == req.Input {
		t.Fatalf("expected transformation for multiline input, got unchanged %q", resp.Escaped)
	}
}
```
