## Expected
- Output contains the `TODO` header.
- Output includes each todo item content.
- Output includes status text so the plan is not just a bare tool heading.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, resp.Output, "TODO")
	assertContains(t, resp.Output, "搜索 Confluence 上 credit.pricing.center 相关文档")
	assertContains(t, resp.Output, "搜索 git 仓库了解项目信息")
	assertContains(t, resp.Output, "in_progress")
}
```