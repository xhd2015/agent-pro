## Expected
- Compile fails
- Error message contains "no Run" and "setup chain"

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Passed {
		t.Fatal("expected compile to fail")
	}
	if !strings.Contains(resp.Output, "no Run") {
		t.Fatalf("expected 'no Run' in output, got:\n%s", resp.Output)
	}
}
```
