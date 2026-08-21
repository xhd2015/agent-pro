## Expected

- No API error.
- Follow-up has open profile and inline `NO_COLOR=1`.
- No `--env-file`; spill dir empty.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	fu := resp.FollowUp
	assertOpenProfile(t, fu, "sess-env-short")
	assertNoEnvFileFlag(t, fu)
	assertContains(t, fu, "NO_COLOR=1")
	if !hasInlineEnvFlag(fu) {
		t.Fatalf("short env must use inline -e; got %q", fu)
	}
	assertSpillDirEmpty(t, resp)
}
```
