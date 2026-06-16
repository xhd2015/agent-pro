## Preconditions
- Priority interactions: multiple detection signals are present simultaneously.
- The higher-priority signal must win.

## Steps
1. Each leaf sets multiple env vars or config fields to test priority ordering.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    // Priority interaction tests — leaves set multiple conflicting env vars
    req.AgentRunnerEnv = ""
    return nil
}
```
