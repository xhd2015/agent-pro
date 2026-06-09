## Expected
- Compile fails
- Error message contains "go block must be final content"

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Passed {
		t.Fatal("expected compile to fail")
	}
	if !strings.Contains(resp.Output, "go block must be final content") {
		t.Fatalf("expected 'go block must be final content' in output, got:\n%s", resp.Output)
	}
}
```
