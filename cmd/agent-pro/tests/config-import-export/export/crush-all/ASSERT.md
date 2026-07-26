## Expected
- Zip file is created with both crush entries under `crush/` prefix.
- Config file goes to `crush/config/crush.json`.
- Data file goes to `crush/data/crush.json`.
- No errors returned.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertZipContains(t, req.ZipPath, []string{
		"crush/config/crush.json",
		"crush/data/crush.json",
	})
}
```
