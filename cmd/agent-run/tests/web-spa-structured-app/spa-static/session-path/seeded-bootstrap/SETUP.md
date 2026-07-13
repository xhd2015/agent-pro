# Scenario

**Feature**: seeded session path injects bootstrap JSON (G3 + G5 accept)

```
seed sessions/spa-seed-bootstrap/ -> GET /sessions/spa-seed-bootstrap
  -> HTML includes #agent-run-session-bootstrap with matching session_id
```

## Preconditions

- Flat session dir exists before web starts so store can load meta/events.

## Steps

1. Seed `spa-seed-bootstrap` under `AGENT_RUN_HOME/sessions/`.
2. Start web; `GET /sessions/spa-seed-bootstrap`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Scenario = "session-path-seeded-bootstrap"
	req.SessionID = "spa-seed-bootstrap"
	if err := seedFlatSession(t, req.Home, req.SessionID, "fake-codex", "idle"); err != nil {
		return err
	}
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}
	req.HTTPPath = "/sessions/" + req.SessionID
	req.HTTPAuth = "none"
	return nil
}
```
