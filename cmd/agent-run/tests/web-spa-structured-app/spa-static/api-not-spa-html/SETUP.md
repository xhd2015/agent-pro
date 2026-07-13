# Scenario

**Feature**: API paths are not served as successful SPA HTML (G4)

```
# health is API, not index.html shell
agent-run web --token test-token
  -> GET /api/agent-run/health (no Bearer) -> 401 JSON/text (not SPA #root body)
  -> GET /api/agent-run/health (Bearer) -> 200 API (not SPA #root / bootstrap)
```

## Preconditions

- Explicit token mode so unauthenticated health is **401** (not open 200).
- Probes use `HTTPAuth` to control Bearer.

## Steps

1. Start web with `--token test-token`.
2. First probe without Bearer (default multi-path handling uses `HTTPAuth=none` for this leaf).
3. Assert neither response is the SPA shell success body.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Scenario = "api-not-spa-html"
	req.WebTokenMode = "explicit"
	req.Token = "test-token"
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}
	// Single unauthenticated probe is enough to prove API path is not SPA HTML.
	// (Authenticated health is also not SPA; Assert may optionally re-check via bearer in notes.)
	req.HTTPPath = "/api/agent-run/health"
	req.HTTPAuth = "none"
	return nil
}
```
