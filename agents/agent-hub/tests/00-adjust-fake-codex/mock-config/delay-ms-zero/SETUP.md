## Preconditions
- The mock config sets `delay_ms` to zero.

## Steps
1. Run fake Codex with two events.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","delay_ms":0,"stdout_events":[{"type":"item.completed","item":{"id":"m1","type":"message","text":"fast one","status":"completed"}},{"type":"item.completed","item":{"id":"m2","type":"message","text":"fast two","status":"completed"}}]}`)
    return nil
}
```

