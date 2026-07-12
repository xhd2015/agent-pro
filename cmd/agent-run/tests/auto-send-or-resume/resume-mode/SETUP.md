# Scenario

**Feature**: MODE=resume when session is bound and exited (resume.ready)

```
seed bound+exited (dead terminal)
  -> auto + followup → provider --resume <runner_session_id>
  -> auto + empty prompt → keep-tty reopen OK
```

## Steps

1. Clear grok-tty command hook so `--agent-runner-binary` argv is visible.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.GrokTTYCommand = ""
	req.Env = withoutEnvKey(req.Env, envGrokTTYCommand)
	req.MetaStatus = "finished"
	return nil
}
```
