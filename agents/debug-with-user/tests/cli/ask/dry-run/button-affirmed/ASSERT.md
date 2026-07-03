## Expected

- Exit code 0.
- Stdout is single-line JSON.
- `answer` = `Yes — window opened`.
- `via` = `button`.
- `affirmed` = `true`.

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	assertJSONField(t, resp, "answer", "Yes — window opened")
	assertJSONField(t, resp, "via", "button")
	assertBoolField(t, resp, "affirmed", true)
}
```
