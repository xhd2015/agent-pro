## Expected

- `--env-file` absolute under spill dir.
- Spill content is the long PATH entry (+ trailing newline).
- Follow-up does not embed the long body; no inline `-e`.

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
	assertOpenProfile(t, fu, "sess-env-long")
	assertHasEnvFileFlag(t, fu)
	assertNoInlineE(t, fu)

	path := resp.EnvFilePath
	if path == "" {
		if p, ok := envFileFlagValue(fu); ok {
			path = p
		}
	}
	assertSpillUnderDir(t, path, req.EnvSpillDir)

	wantBody := req.Env[0] + "\n"
	if resp.SpillFileContent != wantBody {
		t.Fatalf("spill content:\n got %q\nwant %q", resp.SpillFileContent, wantBody)
	}
	if strings.Contains(fu, req.Env[0]) {
		t.Fatalf("follow-up must not embed long env body; line len=%d", len(fu))
	}
	if len(resp.SpillDirEntries) == 0 {
		t.Fatal("expected spill file under EnvSpillDir")
	}
}
```
