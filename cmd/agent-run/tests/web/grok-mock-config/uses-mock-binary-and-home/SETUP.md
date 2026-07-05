# Scenario

**Feature**: web `--grok-home` and `--grok-tty-runner-binary` flow into grok-tty sessions

```
agent-run web --grok-home=<temp> --grok-tty-runner-binary=<wrapper> --agent-runner=grok-tty
POST session runner=grok-tty -> mock wrapper invoked -> grok data under home
```

## Steps

1. Start web with grok mock flags (root helper).
2. POST create session with runner `grok-tty`.
3. `Run` waits for session finished and reads argv probe log.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	startWebWithGrokMock(t, req)
	req.SessionID = postCreateSession(t, req.WebBaseURL, req.WebToken, req.SessionRunner, req.CreatePrompt)
	return nil
}
```