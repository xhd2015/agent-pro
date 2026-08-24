# Scenario

```go
import "time"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeProjectSendSession(t, req)
	req.Procs = nil
	req.OpenFiles = map[int][]string{}
	req.ITerm = nil
	req.AfterOpenHost = false
	req.UseFakeClock = true
	req.OpenWait = 500 * time.Millisecond
	req.Args = []string{"hello", "--session-id", req.SessionID, "--open"}
	return nil
}
```
