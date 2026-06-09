## Expected
- Compile fails
- Error message contains "must define type Request"

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Passed {
		t.Fatal("expected compile to fail")
	}
	if !strings.Contains(resp.Output, "must define type Request") {
		t.Fatalf("expected 'must define type Request' in output, got:\n%s", resp.Output)
	}
}
```
