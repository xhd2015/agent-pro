# Scenario

**Feature**: stateless path still uses agentrunbridge.Run with capture

```
agent.go runAgentStateless
  agentrunbridge.Run + Stateless + CaptureStdout
```

## Steps

1. Mode `stateless_bridge`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "stateless_bridge"
	return nil
}
```
