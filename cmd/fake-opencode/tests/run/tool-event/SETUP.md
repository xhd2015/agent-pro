## Preconditions
- The mock config contains one opencode `tool_use` event.

## Steps
1. Run fake opencode with JSON output.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_tool","stdout_events":[{"type":"tool_use","part":{"id":"t1","type":"tool","tool":"bash","callID":"call_1","state":{"status":"completed","title":"Run tests","output":"ok"}}}]}`)
    return nil
}
```

