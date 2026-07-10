# Scenario

**Feature**: live Socket Mode listener connects with explicit config

```
slack-msg listen --config <repo>/slack-config.json -> connects -> logs config path
```

## Steps

1. Inherit integration grouping setup.
2. No injected events; probe startup + short run.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.InjectEvents = nil
	req.WantAgentCalls = 0
	return nil
}
```
