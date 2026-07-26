## Expected
- Output contains the `SKILL` header.
- Output includes the selected skill name.
- Output includes the command/details supplied in the opencode tool input.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, resp.Output, "SKILL")
	assertContains(t, resp.Output, "git-fetch")
	assertContains(t, resp.Output, "skill install --general-agents")
}
```
