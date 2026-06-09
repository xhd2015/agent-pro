## Expected
- Compile fails
- 3 validation errors collected in one pass
- Root missing Setup or Run, leaf1 ASSERT missing Assert, leaf2 no Run in chain

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Passed {
		t.Fatal("expected compile to fail")
	}
	output := resp.Output
	if !strings.Contains(output, "3 validation errors:") {
		t.Fatalf("expected '3 validation errors' header, got:\n%s", output)
	}
	if !strings.Contains(output, "SETUP.md: must have func Setup or func Run") {
		t.Fatalf("expected root missing Setup or Run error, got:\n%s", output)
	}
	if !strings.Contains(output, "leaf1/ASSERT.md:") {
		t.Fatalf("expected leaf1 ASSERT error, got:\n%s", output)
	}
	if !strings.Contains(output, "missing func Assert") {
		t.Fatalf("expected missing func Assert error, got:\n%s", output)
	}
	if !strings.Contains(output, "no Run") {
		t.Fatalf("expected no Run error, got:\n%s", output)
	}
}
```
