# Scenario

**Feature**: auto-id creates storage session whose id derives from the prompt slug

```
agent-run run --agent-runner fake-codex --session-id-from-prompt "Fix the flaky test"
  -> sessions/fake-codex/fix-the-flaky-test-YYYYMMDD-HHMMSS/
  -> meta.session_id matches directory name
```

## Steps

1. Run with prompt `Fix the flaky test`.
2. Assert single storage session id shape and slug base.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Prompt = "Fix the flaky test"
	req.Args = append(req.Args, req.Prompt)
	return nil
}
```
