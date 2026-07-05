## Expected

- Exit code 0.
- Stdout contains all MVP preset names.
- Codex did not run (`CODEX_HOME=` absent; orchestrator did not start codex foreground).
- Mock server did not announce a listening port in output.

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)

	combined := resp.Stdout + resp.Stderr
	for _, name := range []string{
		"simple",
		"think-message",
		"multi-think",
		"tool-bash",
		"tool-read",
		"think-tool-message",
	} {
		assertContains(t, combined, name)
	}

	assertNotContains(t, combined, "CODEX_HOME=")
	assertNotContains(t, combined, "CODEX_RAN")
}
```