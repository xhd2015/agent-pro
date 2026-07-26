## Expected

- Exit code 0 (L2: `agentruncli.Handle` nil error).
- Stdout contains usage and lists `web`, `run`, `sessions`, `status`, and `--agent-runner`.

```go
import (
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
}
```