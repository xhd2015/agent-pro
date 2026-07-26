## Expected
- Output contains the `WEBFETCH` header.
- Output includes the fetched URL below the header.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, resp.Output, "WEBFETCH")
	assertContains(t, resp.Output, "https://fake.xhd2015.xyz/pages/viewpage.action?pageId=830343951")
}
```