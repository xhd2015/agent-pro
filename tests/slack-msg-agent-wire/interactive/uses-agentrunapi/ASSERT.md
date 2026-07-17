## Expected

- `agent.go` imports or references `agentrunapi` / `pkgs/agentrunapi`.
- `agent.go` references `AutoSendOrResume` (interactive open path).

## Side Effects

- None (source inspection only).

## Errors

- None.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if !resp.HasAgentrunapi {
		t.Fatal("agent.go must import/use github.com/xhd2015/agent-pro/pkgs/agentrunapi")
	}
	if !resp.HasAutoSendOrResume {
		t.Fatal("agent.go must call agentrunapi.AutoSendOrResume for interactive open")
	}
	assertContains(t, resp.AgentSrc, "runAgentInteractiveOpen")
}
```
