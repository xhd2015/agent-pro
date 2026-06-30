# Scenario

**Bug**: `agent-term run grok` fails with `read tcp ... i/o timeout` shortly after attach

Uses real `grok` when on PATH; probes output 4s after attach (past the 2s idle
deadline bug) and expects no timeout error.

```
# grok interactive attach stays alive without WS read timeout
harness PTY -> agent-term run grok -> grok banner visible -> no i/o timeout at 4s
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-pty-probe"
	req.StartDaemon = true
	req.RequireGrok = true
	req.RunProbeSeconds = 4
	req.RunCommand = []string{"grok"}
	return nil
}
```