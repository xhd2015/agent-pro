## Expected

- Exit code 0.
- `answer` = typed user report.
- `via` = `free_text`.
- JSON omits `affirmed` (or it is null — implementer choice; field must not be `true`/`false` for this path).

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	assertJSONField(t, resp, "answer", "VS Code opened but wrong workspace")
	assertJSONField(t, resp, "via", "free_text")
	assertJSONNoKey(t, resp, "affirmed")
}
```
