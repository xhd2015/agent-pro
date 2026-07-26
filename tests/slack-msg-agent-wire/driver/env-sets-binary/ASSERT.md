## Expected

- Source defines / uses env name `SLACK_LISTEN_AGENT_RUN`.
- Source has `agentRunBinary` (or getenv of that env) feeding launch binary /
  `DriverBinary`.

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
	if !resp.HasEnvAgentRun {
		t.Fatal(`agent.go must reference env SLACK_LISTEN_AGENT_RUN`)
	}
	if !resp.HasAgentRunBinary {
		// Allow direct Getenv of the const without helper name only if getenv+const present.
		assertContains(t, resp.AgentSrc, "Getenv")
	} else {
		assertContains(t, resp.AgentSrc, "agentRunBinary")
	}
}
```
