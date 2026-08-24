## Expected

- Validate succeeds; `Update` is `["SINK.md"]`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	ship := assertOK(t, req, resp, err)
	if len(ship.GitCommitFiles.Update) != 1 || ship.GitCommitFiles.Update[0] != "SINK.md" {
		t.Fatalf("update = %+v", ship.GitCommitFiles)
	}
}
```
