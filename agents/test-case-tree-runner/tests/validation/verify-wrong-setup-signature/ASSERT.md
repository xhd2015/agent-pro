## Expected
- Compile fails
- Error message contains "Setup must be func Setup(t *testing.T, req *Request) error"

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Passed {
		t.Fatal("expected compile to fail")
	}
	if !strings.Contains(resp.Output, "Setup must be func Setup(t *testing.T, req *Request) error") {
		t.Fatalf("expected Setup signature error in output, got:\n%s", resp.Output)
	}
}
```
