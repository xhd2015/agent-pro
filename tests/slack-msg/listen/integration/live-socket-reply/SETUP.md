# Scenario

**Feature**: live Socket Mode listener connects with explicit config

```
slack-msg listen --config <repo>/slack-config.json -> connects -> logs config path
```

## Steps

1. Inherit integration grouping setup.
2. No injected events; probe startup + short run.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.InjectEvents = nil
	req.WantAgentCalls = 0
	return nil
}
```
