# Scenario

**Bug**: empty web chat when sync worker never starts without pre-set runner_session_id

```
EnsureGrokSync without GrokSessionID/UpdatesPath
  -> DiscoverBootstrap polls DiscoverSession(prompt, created_at)
  -> delayed grok session dir appears
  -> events.jsonl populated + runner_session_id persisted
```

## Preconditions

- `meta.json` carries `initial_prompt` for discovery bootstrap.
- Grok session dir is **not** present when worker starts (delayed seed).

## Steps

1. Grouping leaves configure discovery bootstrap + delayed grok seed schedule.
2. Assert events and persisted `runner_session_id`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.DiscoveryBootstrap = true
	req.GrokSessionID = ""
	req.UpdatesPath = ""
	return nil
}
```