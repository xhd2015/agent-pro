# Scenario

**Feature**: repeated env flags last-win on the child process

```
run --env A=1 -e A=2 --agent-runner-binary env-logger "prompt"
  -> child A=2
```

## Steps

1. Write env-logging fake runner.
2. Pass both `--env A=1` and `-e A=2` (alias + last-win).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	prepareEnvLoggingRun(t, req)
	req.SessionID = "sess-env-last-win"
	req.Prompt = "env last win"
	req.Args = append(req.Args,
		"--session-id", req.SessionID,
		"--env", "A=1",
		"-e", "A=2",
		"--agent-runner-binary", req.AgentRunnerBinary,
		req.Prompt,
	)
	return nil
}
```
