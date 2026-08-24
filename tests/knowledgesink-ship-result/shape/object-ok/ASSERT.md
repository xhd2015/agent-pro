## Expected

- Validate succeeds.
- `GitCommitFiles.Update` is `["INDEX.md"]`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	ship := assertOK(t, req, resp, err)
	if len(ship.GitCommitFiles.Update) != 1 || ship.GitCommitFiles.Update[0] != "INDEX.md" {
		t.Fatalf("update = %+v", ship.GitCommitFiles)
	}
	if len(ship.GitCommitFiles.Add) != 0 || len(ship.GitCommitFiles.Delete) != 0 {
		t.Fatalf("unexpected buckets: %+v", ship.GitCommitFiles)
	}
}
```
