# Scenario

**Feature**: runAgentStateless may keep agentrunbridge.Run for CaptureStdout

```
runAgentStateless
  -> agentrunbridge.Run(Stateless:true, CaptureStdout:true)
```

## Steps

1. Mode set by leaf.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Mode == "" {
		req.Mode = "stateless_bridge"
	}
	return nil
}
```
