# Scenario

```go
import "github.com/xhd2015/agent-pro/agent/grok/sessions"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeProjectOpenSession(t, req)
	req.Procs = nil
	req.OpenFiles = map[int][]string{}
	req.ITerm = nil
	req.NoAgentRun = true
	req.AgentRunByID = map[string]*sessions.AgentRunOpenResult{
		req.SessionID: {
			AgentRunSessionID: "ar-should-skip",
			Mode:              sessions.AgentRunOpenModeFocus,
			Focused:           true,
		},
	}
	req.Args = []string{req.SessionID, "--no-agent-run"}
	return nil
}
```
