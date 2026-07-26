# Scenario

**Feature**: web-created tty session reuses the same backend terminal across navigation

```
same sessions/<runner>/<session-id>/meta.json
same <runner>-registry/<session-id>.json
GET /terminal before navigation -> available terminal id
GET session detail -> return to same chat -> GET /terminal -> same terminal id
```

## Preconditions

- The tty registry session id matches the chat session id or is persistently mapped.
- Navigating away from the browser page does not spawn a new backend tty.

## Steps

1. Write one tty session and one live registry.
2. `Run` fetches terminal availability for the existing session.
3. `Assert` fetches it again and compares the resolved identity.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "http"
	req.Runner = "codex-tty"
	req.SessionID = "web-created-tty-session"
	req.RegistryTranscript = "durable-terminal\n"
	listenAddr := startFakePtywrap(t, req)
	writeSessionFixture(t, req, req.Runner, req.SessionID, "running")
	writeTTYRegistryFixture(t, req, req.Runner, req.SessionID, listenAddr)
	req.HTTPPath = terminalStatusPath(req.Runner, req.SessionID)
	return nil
}
```
