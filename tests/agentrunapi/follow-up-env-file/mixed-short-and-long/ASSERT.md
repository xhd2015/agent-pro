## Expected

- One `--env-file`; no inline `-e`.
- Spill file contains short and long entries (all three lines).

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
	assertOpenProfile(t, fu, "sess-env-mixed")
	assertHasEnvFileFlag(t, fu)
	assertNoInlineE(t, fu)

	path := resp.EnvFilePath
	if path == "" {
		if p, ok := envFileFlagValue(fu); ok {
			path = p
		}
	}
	assertSpillUnderDir(t, path, req.EnvSpillDir)

	want := "NO_COLOR=1\n" + longPATHEntry() + "\nFOO=bar\n"
	if resp.SpillFileContent != want {
		t.Fatalf("spill must contain ALL entries:\n got %q\nwant %q", resp.SpillFileContent, want)
	}
	for _, e := range req.Env {
		if strings.Contains(fu, e) && strings.HasPrefix(e, "PATH=") {
			t.Fatalf("long PATH must not appear on follow-up line")
		}
	}
}
```
