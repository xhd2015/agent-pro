## Preconditions
- The mock config contains stderr text and exit code zero.

## Steps
1. Run fake Codex with the mock config.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","stderr":"diagnostic line","exit_code":0,"llm_events":[]}`)
    return nil
}
```

