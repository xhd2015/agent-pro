# Scenario

**Feature**: grok mock web harness creates and finishes a grok-tty session

```
startWebWithGrokMock -> POST grok-tty -> wait finished
```

## Steps

1. Start web with grok mock flags.
2. POST create session with runner `grok-tty`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	startWebWithGrokMock(t, req)
	req.SessionID = postCreateSession(t, req.WebBaseURL, req.WebToken, req.SessionRunner, req.CreatePrompt)
	return nil
}
```