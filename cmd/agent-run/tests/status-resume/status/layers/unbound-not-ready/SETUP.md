# Scenario

**Feature**: meta without runner_session_id ⇒ runner unbound, resume not ready

```
meta without runner_session_id
  -> agent-run status test-unbound-s1
  -> runner.status unbound, resume.ready no
```

## Steps

1. Seed unbound meta.
2. Run human status.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "test-unbound-s1"
	req.MetaStatus = "finished"
	req.InitialPrompt = "never bound"
	seedUnbound(t, req)
	req.Args = []string{"status", req.SessionID}
	return nil
}
```
