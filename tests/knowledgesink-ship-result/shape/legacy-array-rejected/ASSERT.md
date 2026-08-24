## Expected

- Error mentions that `git_commit_files` must be an object.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrContains(t, req, resp, err)
}
```
