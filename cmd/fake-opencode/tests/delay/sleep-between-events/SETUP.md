## Preconditions
- Mock config has three events: message "before", sleep(500ms), message "after".

## Steps
1. Write mock config with a sleep event between two message events.
2. Run fake-opencode.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_sleep","stdout_events":[{"type":"message","text":"before"},{"type":"sleep","delay_ms":500},{"type":"message","text":"after"}]}`)
    return nil
}
```
