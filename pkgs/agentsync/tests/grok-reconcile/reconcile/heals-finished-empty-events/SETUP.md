# Scenario

**Bug**: R2 — reconcile heals finished session with empty events.jsonl

```
meta finished, no events.jsonl, grok updates pre-seeded
  -> ReconcileOnce(session)
  -> events.jsonl has user + assistant from grok updates
```

## Steps

1. Seed `meta.json` with `initial_prompt`, `status=finished`, empty `runner_session_id`.
2. Pre-seed grok `updates.jsonl` with matching prompt + turn completion.
3. Call `ReconcileOnce`; poll for events.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "reconcile-heal"
	req.SessionID = "reconcile-heal-test"
	req.InitialPrompt = reconcileHealPrompt
	req.GrokSessionID = reconcileHealGrokUUID
	req.SessionStatus = "finished"
	return nil
}
```