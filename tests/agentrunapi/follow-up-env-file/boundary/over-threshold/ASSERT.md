## Expected

- `--env-file` present; no inline `-e`; spill body matches entry.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	fu := resp.FollowUp
	assertOpenProfile(t, fu, "sess-env-over")
	assertHasEnvFileFlag(t, fu)
	assertNoInlineE(t, fu)
	path := resp.EnvFilePath
	if path == "" {
		if p, ok := envFileFlagValue(fu); ok {
			path = p
		}
	}
	assertSpillUnderDir(t, path, req.EnvSpillDir)
	want := req.Env[0] + "\n"
	if resp.SpillFileContent != want {
		t.Fatalf("spill:\n got %q\nwant %q", resp.SpillFileContent, want)
	}
	if strings.Contains(fu, req.Env[0]) {
		t.Fatalf("over-threshold entry must not appear on follow-up line")
	}
}
```
