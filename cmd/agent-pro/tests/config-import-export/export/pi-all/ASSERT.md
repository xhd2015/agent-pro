## Expected
- Zip file is created with all pi entries under `pi/` prefix.
- No errors returned.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertZipContains(t, req.ZipPath, []string{
		"pi/auth.json",
		"pi/settings.json",
		"pi/models.json",
	})
}
```
