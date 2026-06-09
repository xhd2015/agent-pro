## Expected
- Compile fails
- Error message contains "Run must be func Run(t *testing.T, req *Request) (*Response, error)"

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Passed {
		t.Fatal("expected compile to fail")
	}
	if !strings.Contains(resp.Output, "Run must be func Run(t *testing.T, req *Request) (*Response, error)") {
		t.Fatalf("expected Run signature error in output, got:\n%s", resp.Output)
	}
}
```
