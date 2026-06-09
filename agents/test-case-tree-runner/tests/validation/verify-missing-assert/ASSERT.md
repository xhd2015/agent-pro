## Expected
- Compile fails
- Error message contains "missing func Assert" and references leaf/ASSERT.md

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Passed {
		t.Fatal("expected compile to fail")
	}
	if !strings.Contains(resp.Output, "missing func Assert") {
		t.Fatalf("expected 'missing func Assert' in output, got:\n%s", resp.Output)
	}
	if !strings.Contains(resp.Output, "leaf/ASSERT.md") {
		t.Fatalf("expected leaf ASSERT.md reference in output, got:\n%s", resp.Output)
	}
}
```
