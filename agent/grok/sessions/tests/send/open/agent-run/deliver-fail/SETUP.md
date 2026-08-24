# Scenario

Managed id but deliver fails → hard error; no bare ForceNew.

```go
import "fmt"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeProjectSendSession(t, req)
	req.Procs = nil
	req.OpenFiles = map[int][]string{}
	req.ITerm = nil
	req.AfterOpenHost = true // would enable bare resume if soft-missed
	req.AgentRunErr = fmt.Errorf("terminal unreachable at 127.0.0.1:9")
	req.Args = []string{"hello", "--session-id", req.SessionID, "--open"}
	return nil
}
```
