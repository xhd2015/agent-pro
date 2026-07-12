# Scenario

**Feature**: writer close code 4000 deletes session and kills child

Control path already used by the React client for unused StrictMode sessions
(`ws.close(4000)`). Must remove the session from the list and free the PTY.

```
WS attach as writer -> close 4000 -> !ProcessAlive && !SessionListed
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "lifecycle-writer-close"
	req.WSCloseCode = 4000
	return nil
}
```
