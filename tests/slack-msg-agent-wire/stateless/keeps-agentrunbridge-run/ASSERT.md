## Expected

- `runAgentStateless` still present.
- Still uses `agentrunbridge` for this path (`agentrunbridge.Run`).
- Mentions `Stateless` and `CaptureStdout` (stdout capture for PostMessage).

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
	assertContains(t, resp.AgentSrc, "runAgentStateless")
	if !resp.HasAgentrunbridge {
		t.Fatal("stateless path may keep agentrunbridge import")
	}
	if !resp.HasBridgeRun {
		t.Fatal("runAgentStateless must call agentrunbridge.Run for CaptureStdout")
	}
	if !resp.HasStateless {
		t.Fatal("stateless path must set Stateless")
	}
	if !resp.HasCaptureStdout {
		t.Fatal("stateless path must set CaptureStdout")
	}
}
```
