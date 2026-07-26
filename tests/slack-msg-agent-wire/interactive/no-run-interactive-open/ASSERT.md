## Expected

- `agent.go` does **not** reference `RunInteractiveOpen`.
- Interactive entrypoint `runAgentInteractiveOpen` still exists (name may stay).

## Side Effects

- None.

## Errors

- None.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.HasRunInteractiveOpen {
		t.Fatal("agent.go must not call agentrunbridge.RunInteractiveOpen after P4 cutover")
	}
	assertContains(t, resp.AgentSrc, "runAgentInteractiveOpen")
}
```
