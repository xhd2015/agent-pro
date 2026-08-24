## Expected

- Validate succeeds; `Add` is `["topics/new.md"]`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	ship := assertOK(t, req, resp, err)
	if len(ship.GitCommitFiles.Add) != 1 || ship.GitCommitFiles.Add[0] != "topics/new.md" {
		t.Fatalf("add = %+v", ship.GitCommitFiles)
	}
}
```
