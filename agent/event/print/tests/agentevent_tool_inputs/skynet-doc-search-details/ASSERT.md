## Expected
- Output contains the skynet tool header.
- Output includes the Confluence search URL below the header.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, resp.Output, "SKYNET-BASE_GET_DOC_CONTENT")
	assertContains(t, resp.Output, "https://fake.xhd2015.xyz/search?text=credit+pricing+center")
}
```