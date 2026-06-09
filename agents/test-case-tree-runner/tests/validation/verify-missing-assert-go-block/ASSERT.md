## Expected
- Compile fails
- Error message contains "missing go block" from ASSERT.md parsing

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Passed {
		t.Fatal("expected compile to fail")
	}
	if !strings.Contains(resp.Output, "missing go block") {
		t.Fatalf("expected missing go block error, got:\n%s", resp.Output)
	}
	if !strings.Contains(resp.Output, "ASSERT.md") {
		t.Fatalf("expected ASSERT.md reference, got:\n%s", resp.Output)
	}
}
```
