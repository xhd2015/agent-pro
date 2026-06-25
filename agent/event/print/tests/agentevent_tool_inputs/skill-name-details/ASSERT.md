## Expected
- Output contains the `SKILL` header.
- Output includes the selected skill name below the header.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, resp.Output, "SKILL")
	assertContains(t, resp.Output, "confluence-fetch")
}
```