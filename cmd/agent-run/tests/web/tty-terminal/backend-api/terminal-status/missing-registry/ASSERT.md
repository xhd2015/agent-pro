## Expected

- HTTP 200.
- JSON does not advertise attach availability.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertHTTPStatus(t, resp, 200)
	obj := decodeJSONBody(t, resp.HTTPBody)
	if boolField(obj, "available") {
		t.Fatalf("missing registry must not be attachable: %s", resp.HTTPBody)
	}
}
```
