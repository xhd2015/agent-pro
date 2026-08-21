## Expected

- Follow-up contains `--env-file` pointing at the given path.
- EnvSpillDir has no auto-spill files.
- No inline `-e`.

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	fu := resp.FollowUp
	assertOpenProfile(t, fu, "sess-env-explicit")
	assertHasEnvFileFlag(t, fu)
	assertNoInlineE(t, fu)

	path := resp.EnvFilePath
	if path == "" {
		if p, ok := envFileFlagValue(fu); ok {
			path = p
		}
	}
	assertAbsPath(t, path)
	wantAbs, _ := filepath.Abs(req.EnvFile)
	if path != wantAbs && !strings.HasSuffix(path, "given/path.env") {
		t.Fatalf("env-file path=%q want given %q", path, wantAbs)
	}
	assertSpillDirEmpty(t, resp)
}
```
