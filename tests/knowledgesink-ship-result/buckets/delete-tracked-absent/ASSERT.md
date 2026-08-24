## Expected

- Validate succeeds; `Delete` is `["README.md"]`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	ship := assertOK(t, req, resp, err)
	if len(ship.GitCommitFiles.Delete) != 1 || ship.GitCommitFiles.Delete[0] != "README.md" {
		t.Fatalf("delete = %+v", ship.GitCommitFiles)
	}
}
```
