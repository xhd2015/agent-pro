# Scenario

**Feature**: DiscoverSession matches nested envelope first user chunk

```
seed session dir with envelope user_message_chunk -> DiscoverSession(prompt) -> correct uuid
```

## Preconditions

- `updates.jsonl` first user chunk is nested `_x.ai/session/update` envelope.
- Prompt `run ls` must match extracted text via `grok_session.ParseLine`.

## Steps

1. Seed grok session dir with envelope user chunk and recent `created_at`.
2. Set `req.Prompt = "run ls"` and call `DiscoverSession`.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.GrokHome = t.TempDir()
	req.Workspace = t.TempDir()
	req.SessionUUID = "33333333-3333-3333-3333-333333333333"
	req.Prompt = "run ls"
	req.RunStart = time.Now().Add(-time.Second)
	seedGrokSessionDir(
		t,
		req.GrokHome,
		req.Workspace,
		req.SessionUUID,
		req.Prompt,
		time.Now().UTC(),
		acpEnvelopeUserChunk(req.SessionUUID, "run ls"),
	)
	return nil
}
```