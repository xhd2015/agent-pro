# Scenario

```go
import "github.com/xhd2015/agent-pro/agent/grok/sessions"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeProjectSendSession(t, req)
	req.Procs = nil
	req.OpenFiles = map[int][]string{}
	req.ITerm = nil
	req.AgentRunByID = map[string]*sessions.AgentRunOpenResult{
		req.SessionID: {
			AgentRunSessionID: "ar-resume-1",
			Mode:              sessions.AgentRunOpenModeResume,
			Opened:            true,
			Delivered:         true,
			CWD:               req.ProjectDir,
			Command:           "/usr/local/bin/agent-run run --session-id ar-resume-1 --auto-send-or-resume --open -- hello",
		},
	}
	req.Args = []string{"hello", "--session-id", req.SessionID, "--open"}
	return nil
}
```
