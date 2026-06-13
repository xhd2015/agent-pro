## Preconditions
- The mock config contains only text and error events (no tool_use).

## Steps
1. Write a mock config with one text event and one error event.
2. Run fake-opencode and verify these events pass through unchanged.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_nontool","stdout_events":[{"type":"text","part":{"id":"p1","type":"text","text":"plain text message"}},{"type":"error","error":{"name":"TestError","data":{"message":"an error occurred"}}}]}`)
    return nil
}
```
