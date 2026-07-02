## Expected

- HTTP 200 or another clear non-terminal status.
- It must not advertise `available: true`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.HTTPStatus != 200 && resp.HTTPStatus != 404 {
		t.Fatalf("unexpected non-tty terminal status=%d body=%s", resp.HTTPStatus, resp.HTTPBody)
	}
	if resp.HTTPStatus == 200 {
		obj := decodeJSONBody(t, resp.HTTPBody)
		if boolField(obj, "available") {
			t.Fatalf("non-tty runner must not advertise terminal: %s", resp.HTTPBody)
		}
	}
}
```
