# Scenario

**Bug**: `agent-term serve` stderr logs mash listen address into request lines

Reproduces `127.0.0.1:7681POST /api/terminal/sessions` missing newline between
startup banner and request log.

```
# serve logs
agent-term serve -> stderr startup line -> POST create -> each log on its own line
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "serve-logs-on-create"
	req.StartDaemon = true
	return nil
}
```