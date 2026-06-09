## Expected
- Compile fails
- Error message contains "Assert must be func Assert(t *testing.T, req *Request, resp *Response, err error)"

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Passed {
		t.Fatal("expected compile to fail")
	}
	if !strings.Contains(resp.Output, "Assert must be func Assert(t *testing.T, req *Request, resp *Response, err error)") {
		t.Fatalf("expected Assert signature error in output, got:\n%s", resp.Output)
	}
}
```
