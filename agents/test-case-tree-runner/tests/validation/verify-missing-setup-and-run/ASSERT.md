## Expected
- Compile fails
- Error message contains "must have func Setup or func Run"

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Passed {
		t.Fatal("expected compile to fail")
	}
	if !strings.Contains(resp.Output, "must have func Setup or func Run") {
		t.Fatalf("expected Setup or Run error, got:\n%s", resp.Output)
	}
}
```
