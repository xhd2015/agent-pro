# Scenario

**Feature**: DiscoverSession matches prompt via grok_session.ParseLine

```
# seed grok session dir with summary.json + updates.jsonl
GROK_HOME/sessions/<encoded-cwd>/<uuid>/ -> DiscoverSession(cwd, prompt) -> matching session id
```

## Preconditions

- Session `summary.json` cwd matches workspace; `created_at` is after `runStart`.
- First `user_message_chunk` in `updates.jsonl` (flat or envelope) equals prompt.

## Steps

1. Set `req.Target = "session"`.
2. Leaf SETUPs seed session dir under temp `GROK_HOME`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Target = "session"
	return nil
}
```