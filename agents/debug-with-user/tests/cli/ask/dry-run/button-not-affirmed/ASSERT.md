## Expected

- Exit code 0.
- `answer` = `No — window did not open`.
- `via` = `button`.
- `affirmed` = `false`.

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	assertJSONField(t, resp, "answer", "No — window did not open")
	assertJSONField(t, resp, "via", "button")
	assertBoolField(t, resp, "affirmed", false)
}
```
