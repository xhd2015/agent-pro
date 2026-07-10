## Expected

- `ready=false`, `state=unknown`, `reason` contains `no terminal output`.

## Exit Code

N/A (direct package call)

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertWritable(t, "empty snapshot", resp.Status, false, "unknown", "no terminal output")
}
```
