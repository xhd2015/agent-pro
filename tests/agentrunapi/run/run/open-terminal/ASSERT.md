## Expected

- No API error.
- OpenFn called once.
- Follow-up contains `--open`, `AGENT_RUN_HOME=`, and StoreHome.
- Follow-up has no `--new-terminal` and no `--detach`.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	assertEqual(t, "OpenCalls", resp.OpenCalls, 1)
	assertContains(t, resp.OpenFollowUp, "--open")
	assertContains(t, resp.OpenFollowUp, "AGENT_RUN_HOME=")
	assertContains(t, resp.OpenFollowUp, req.StoreHome)
	assertNotContains(t, resp.OpenFollowUp, "--new-terminal")
	if strings.Contains(resp.OpenFollowUp, "--detach") {
		t.Fatalf("open follow-up should not use --detach: %s", resp.OpenFollowUp)
	}
}
```
