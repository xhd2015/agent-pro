## Expected
- An error is returned because the zip file cannot be created.
- The zip file does not exist on disk.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertError(t, resp)
	entries := readZipEntries(t, req.ZipPath)
	if entries != nil {
		t.Fatal("zip file should not exist")
	}
}
```
