## Preconditions
- The mock config contains a nonzero exit code.

## Steps
1. Run fake Codex with the mock config.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","stderr":"planned failure","exit_code":7,"stdout_events":[{"type":"item.completed","item":{"id":"m1","type":"message","text":"before failure","status":"completed"}}]}`)
    return nil
}
```

