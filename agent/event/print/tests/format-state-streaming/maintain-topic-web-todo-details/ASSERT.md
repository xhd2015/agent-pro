## Expected
- Confluence search, todo plan, and web fetch blocks each include their input details.
- Bare headers without URL or todo content are insufficient.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, resp.Output, "SKYNET-BASE_GET_DOC_CONTENT")
	assertContains(t, resp.Output, "https://fake.xhd2015.xyz/search?text=credit+pricing+center")
	assertContains(t, resp.Output, "TODO")
	assertContains(t, resp.Output, "搜索 Confluence 上 credit.pricing.center 相关文档")
	assertContains(t, resp.Output, "WEBFETCH")
	assertContains(t, resp.Output, "https://fake.xhd2015.xyz/pages/viewpage.action?pageId=830343951")
}
```