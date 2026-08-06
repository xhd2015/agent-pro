## Expected

- Exit code 0 (L2: `agentruncli.Handle` nil error).
- Stdout contains usage and lists `web`, `run`, `sessions`, `status`, `kill`, and `--agent-runner`.
- `kill` appears as a top-level command inventory line (not only inside `pty` “kill orphan”).

```go
import (
	"regexp"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertExitCode(t, resp, 0)
	assertOutput(t, resp, "stdout",
		"Usage:",
		"web",
		"run",
		"sessions",
		"status",
		"--agent-runner",
	)
	// Require kill as its own command line; avoid matching "kill orphan" under pty.
	cmdLine := regexp.MustCompile(`(?m)^[ \t]+kill\b`)
	if !cmdLine.MatchString(resp.Stdout) {
		t.Fatalf("top-level --help must list kill as a command, got:\n%s", resp.Stdout)
	}
}
```
