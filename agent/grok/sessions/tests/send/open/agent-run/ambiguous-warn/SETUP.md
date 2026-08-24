# Scenario

```go
import "fmt"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeProjectSendSession(t, req)
	req.Procs = nil
	req.OpenFiles = map[int][]string{}
	req.ITerm = nil
	req.AfterOpenHost = true
	req.AgentRunErr = fmt.Errorf("ambiguous grok-session-id %s: multiple matches: ar-a, ar-b", req.SessionID)
	req.Args = []string{"hello", "--session-id", req.SessionID, "--open"}
	return nil
}
```
