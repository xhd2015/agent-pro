# Scenario

**Feature**: SLACK_LISTEN_AGENT_RUN env selects driver binary

```
const envAgentRun = "SLACK_LISTEN_AGENT_RUN"
agentRunBinary() reads that env
```

## Steps

1. Mode `driver_env`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "driver_env"
	return nil
}
```
