## Preconditions
- The mock config contains a sequence: think, tool_call, message.

## Steps
1. Run fake opencode with the mock config.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_multi","llm_events":[{"type":"think","text":"initial analysis"},{"type":"tool_call","tool":"bash","tool_input":{"command":"echo mid command"}},{"type":"message","text":"final summary"}]}`)
    return nil
}
```
