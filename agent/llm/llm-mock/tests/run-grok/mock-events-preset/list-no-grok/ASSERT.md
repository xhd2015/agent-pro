## Expected

- Exit code 0.
- Stdout contains all MVP preset names.
- Grok did not run (`GROK_HOME=` absent; orchestrator did not start grok foreground).
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

	assertNotContains(t, combined, "GROK_HOME=")
	assertNotContains(t, combined, "GROK_RAN")
}
```