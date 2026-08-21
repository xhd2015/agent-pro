## Expected

- No `--env-file`; entry appears inline; spill dir empty.

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
	assertOpenProfile(t, fu, "sess-env-at")
	assertNoEnvFileFlag(t, fu)
	assertContains(t, fu, req.Env[0])
	assertSpillDirEmpty(t, resp)
}
```
