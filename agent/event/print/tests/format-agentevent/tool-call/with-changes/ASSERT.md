## Expected
- Output contains the EDIT icon/label for write tool.
- Output contains the file change `create src/main.go`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, resp.Output, "EDIT")
	assertContains(t, resp.Output, "create")
	assertContains(t, resp.Output, "src/main.go")
}
```
