# Scenario

**Feature**: `status --json` mirrors multi-layer shape including runner.exited and resume.ready

```
seed bound+exited meta
  -> agent-run status --json test-json-s1
  -> JSON runner.exited + resume.ready
```

## Steps

1. Seed bound exited meta.
2. Run status with `--json`.
3. Mode `status-json` parses stdout.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "test-json-s1"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440333"
	req.MetaStatus = "finished"
	req.TerminalSessionID = "term-json-1"
	req.InitialPrompt = "json shape"
	req.WriteRegistry = false
	seedBoundExitedDeadTerminal(t, req)
	req.Mode = "status-json"
	req.Args = []string{"status", "--json", req.SessionID}
	return nil
}
```
