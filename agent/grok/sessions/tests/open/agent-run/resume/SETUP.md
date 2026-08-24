# Scenario

```go
import "github.com/xhd2015/agent-pro/agent/grok/sessions"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeProjectOpenSession(t, req)
	req.Procs = nil
	req.OpenFiles = map[int][]string{}
	req.ITerm = nil
	req.AgentRunByID = map[string]*sessions.AgentRunOpenResult{
		req.SessionID: {
			AgentRunSessionID: "ar-open-resume",
			Mode:              sessions.AgentRunOpenModeResume,
			Opened:            true,
			CWD:               req.ProjectDir,
			Command:           "/usr/local/bin/agent-run run --session-id ar-open-resume --auto-send-or-resume --open",
		},
	}
	return nil
}
```
