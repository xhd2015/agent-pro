# Scenario

```go
import "github.com/xhd2015/agent-pro/agent/grok/sessions"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeProjectSendSession(t, req)
	req.Procs = nil
	req.OpenFiles = map[int][]string{}
	req.ITerm = nil
	req.AfterOpenHost = true
	req.NoAgentRun = true
	req.AgentRunByID = map[string]*sessions.AgentRunOpenResult{
		req.SessionID: {
			AgentRunSessionID: "ar-should-skip",
			Mode:              sessions.AgentRunOpenModeSend,
			Delivered:         true,
		},
	}
	req.Args = []string{"hello", "--session-id", req.SessionID, "--open", "--no-agent-run"}
	return nil
}
```
