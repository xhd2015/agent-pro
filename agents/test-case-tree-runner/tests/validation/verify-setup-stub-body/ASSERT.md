## Expected
- Compile fails
- Error message contains "func Setup body must not be a stub"

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Passed {
		t.Fatal("expected compile to fail")
	}
	if !strings.Contains(resp.Output, "func Setup body must not be a stub") {
		t.Fatalf("expected 'func Setup body must not be a stub' in output, got:\n%s", resp.Output)
	}
}
```
