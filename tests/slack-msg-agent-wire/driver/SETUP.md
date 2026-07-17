# Scenario

**Feature**: DriverBinary / SLACK_LISTEN_AGENT_RUN policy

```
agentRunBinary() = getenv(SLACK_LISTEN_AGENT_RUN)
  empty -> agent-run default (library DriverBinary "")
  set   -> that binary
```

## Steps

1. Mode set per leaf.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.Mode == "" {
		req.Mode = "driver_empty"
	}
	return nil
}
```
