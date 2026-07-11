# Scenario

**Feature**: resume denied when runner_session_id is not bound

```
meta without runner_session_id -> resume <id> "x" -> exit 1, not bound
```

## Steps

1. Seed unbound meta.
2. Run resume with followup.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "test-resume-unbound-s1"
	req.MetaStatus = "finished"
	req.InitialPrompt = "unbound prior"
	seedUnbound(t, req)
	req.Args = []string{"resume", req.SessionID, "followup"}
	return nil
}
```
