## Expected
- Compile fails
- Error message contains "must have a Go code block"

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Passed {
		t.Fatal("expected compile to fail")
	}
	if !strings.Contains(resp.Output, "must have a Go code block") {
		t.Fatalf("expected Go code block error, got:\n%s", resp.Output)
	}
}
```
