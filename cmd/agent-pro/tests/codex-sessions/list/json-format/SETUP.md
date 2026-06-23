# Scenario

**Feature**: list JSON output matches FormatListJSON schema

```
# one rollout session
writeRolloutSession -> sessions.List -> FormatListJSON

# JSON object with sessions array and metadata fields
{"sessions":[{"id","started_at","cwd","path"}]}
```

## Preconditions

- `req.Format = "json"` selects JSON formatting.

## Steps

1. Create session `01900005-bbbb-7bbb-bbbb-bbbbbbbbbbbb`.
2. Set `req.Format = "json"`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Limit = 10
	req.Format = "json"
	writeRolloutSession(t, req.CodexHome,
		"01900005-bbbb-7bbb-bbbb-bbbbbbbbbbbb",
		"2026-06-23T11:00:00.000Z", "/tmp/project-b")
	return nil
}
```