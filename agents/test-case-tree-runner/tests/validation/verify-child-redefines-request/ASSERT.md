## Expected
- Compile fails
- Error message contains "cannot redefine Request"

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Passed {
		t.Fatal("expected compile to fail")
	}
	if !strings.Contains(resp.Output, "cannot redefine Request") {
		t.Fatalf("expected 'cannot redefine Request' in output, got:\n%s", resp.Output)
	}
}
```
