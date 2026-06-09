```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Result != 10 {
		t.Fatalf("expected 10, got %d", resp.Result)
	}
}
```
