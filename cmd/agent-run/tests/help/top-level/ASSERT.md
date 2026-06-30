## Expected

- Exit code 0.
- Stdout contains usage and lists `web`, `run`, `sessions`, `status`, and `--agent-runner`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
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