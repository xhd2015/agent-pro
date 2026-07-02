## Expected

- HTTP 200.
- JSON reports unavailable and does not leak local registry paths.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertHTTPStatus(t, resp, 200)
	obj := decodeJSONBody(t, resp.HTTPBody)
	if boolField(obj, "available") {
		t.Fatalf("stale registry must not be attachable: %s", resp.HTTPBody)
	}
	requireNoPathLeak(t, resp.HTTPBody)
}
```
