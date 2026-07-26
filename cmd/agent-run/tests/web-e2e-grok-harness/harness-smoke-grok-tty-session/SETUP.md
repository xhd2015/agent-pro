# Scenario

**Feature**: grok mock web harness creates and finishes a grok-tty session

```
startWebWithGrokMock -> POST grok-tty -> wait finished
```

## Steps

1. Start web with grok mock flags.
2. POST create session with runner `grok-tty`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	startWebWithGrokMock(t, req)
	req.SessionID = postCreateSession(t, req.WebBaseURL, req.WebToken, req.SessionRunner, req.CreatePrompt)
	return nil
}
```