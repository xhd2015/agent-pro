## Preconditions
- The mock config uses `stdout_events` with native codex format (has `item` field).
- Backward compatibility: native codex format auto-detection in `stdout_events` must still work.

## Steps
1. Run fake codex with the mock config.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","stdout_events":[{"type":"item.started","item":{"id":"1","type":"reasoning","text":"native format backcompat"}}]}`)
    return nil
}
```
