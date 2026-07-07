# Scenario

**Feature**: web run stores assistant events without phased streaming rows

```
POST fake-codex session -> wait finished -> inspect events.jsonl
```

## Steps

1. Start web; create fake-codex session.
2. Wait until finished; read persisted events.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	startAgentRunWeb(t, req)
	req.Runner = "fake-codex"
	req.Prompt = "no phases please"
	sessionID, _, _ := postCreateSession(t, req, req.Runner, req.Prompt)
	req.SessionID = sessionID
	waitForSessionStatus(t, req, req.Runner, sessionID, "finished", 45*time.Second)
	req.Mode = "cli"
	req.CLIArgs = []string{"sessions", req.Runner + "/" + sessionID, "--print"}
	return nil
}
```
