# Scenario

**Bug**: A6 — discovery bootstrap heals delayed grok session (PRIMARY empty-chat fix)

```
no grok session at worker start (empty GROK_HOME)
  -> agentsync.EnsureGrokSync with InitialPrompt only
  -> delayed updates.jsonl + turn completion 600ms later
  -> events emitted; runner_session_id persisted to meta.json
```

Mirrors production bug `web_a5939cab5f4c7bfe`: grok completed turn ~2s after
agent-run marked finished; worker must discover session without pre-set id.

## Steps

1. Write `meta.json` with `initial_prompt`; leave `runner_session_id` empty.
2. Start `EnsureGrokSync` without `GrokSessionID` / `UpdatesPath`.
3. Schedule delayed grok session dir + turn completion after 600ms.
4. Poll until user + assistant events appear.

```go
import (
	"testing"
	"time"
)

const (
	discoverBootstrapPrompt   = "run ls and pwd"
	discoverBootstrapGrokUUID = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	discoverBootstrapReply    = "discover-bootstrap-reply-marker"
)

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "discover-bootstrap-test"
	req.InitialPrompt = discoverBootstrapPrompt
	req.SessionCreatedAt = time.Now().Add(-1 * time.Second)
	req.DiscoveryBootstrap = true
	req.GrokSessionID = ""
	req.UpdatesPath = ""
	req.DelayedGrokSeed = &DelayedGrokSeed{
		Delay:         600 * time.Millisecond,
		GrokSessionID: discoverBootstrapGrokUUID,
		Prompt:        discoverBootstrapPrompt,
		Lines: []string{
			acpAgentMessageChunk(discoverBootstrapReply),
			acpTurnCompleted(),
		},
	}
	req.DiscoveryTimeout = 10 * time.Second
	req.HoldAfterSchedule = 0
	return nil
}
```