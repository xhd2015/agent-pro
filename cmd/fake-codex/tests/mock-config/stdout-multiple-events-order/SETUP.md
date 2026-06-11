## Preconditions
- The mock config contains two message events in a specific order.

## Steps
1. Run fake Codex with the mock config.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","stdout_events":[{"type":"item.completed","item":{"id":"m1","type":"message","text":"first response","status":"completed"}},{"type":"item.completed","item":{"id":"m2","type":"message","text":"second response","status":"completed"}}]}`)
    return nil
}
```

