## Preconditions
- The command runs with a temporary HOME.

## Steps
1. Run fake opencode with an empty mock config.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","stdout_events":[]}`)
    return nil
}
```

