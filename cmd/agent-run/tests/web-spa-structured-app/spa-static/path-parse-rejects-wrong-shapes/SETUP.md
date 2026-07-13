# Scenario

**Feature**: `parseSessionPagePath` rejects wrong shapes — no bootstrap (G5)

```
# seed exists for id "spa-parse-id" but only exact /sessions/spa-parse-id injects
seed spa-parse-id
  GET /sessions/                    -> SPA ok, no bootstrap
  GET /sessions                     -> SPA ok, no bootstrap
  GET /sessions/spa-parse-id/extra  -> SPA ok, no bootstrap
  GET /session/spa-parse-id         -> SPA ok, no bootstrap
  # note: /sessions//id is URL-cleaned by the HTTP stack to /sessions/id — not a reject case
  GET /sessions/spa-parse-id        -> (positive control covered by seeded-bootstrap leaf)
```

## Preconditions

- Seeded session `spa-parse-id` so a correct path *would* bootstrap; wrong shapes must still omit inject.

## Steps

1. Seed flat session `spa-parse-id`.
2. Start web.
3. Multi-GET wrong shapes; Assert none include bootstrap.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Scenario = "path-parse-rejects"
	req.SessionID = "spa-parse-id"
	if err := seedFlatSession(t, req.Home, req.SessionID, "fake-codex", "idle"); err != nil {
		return err
	}
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}
	req.HTTPAuth = "none"
	req.HTTPPaths = []string{
		"/sessions/",
		"/sessions",
		"/sessions/" + req.SessionID + "/extra",
		"/session/" + req.SessionID,
	}
	return nil
}
```
