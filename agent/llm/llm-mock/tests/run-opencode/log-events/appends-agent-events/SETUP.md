# Scenario

**Feature**: `--log-events` with `--mock-events-preset=simple` appends message AgentEvent

```
llm-mock run --mock-events-preset=simple --log-events f.jsonl opencode
orchestrator -> mock genQueue [message]
fake opencode -> curl /v1/chat/completions once -> log AgentEvents
```

## Steps

1. No config env (default empty `exchanges[]`).
2. `--mock-events-preset=simple` and `--log-events` for AgentEvent proof.
3. Fake opencode curls mock once via baseURL from `$OPENCODE_CONFIG_CONTENT`.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ConfigJSON = ""
	req.ConfigEnv = ""
	req.MockEventsPreset = "simple"
	req.LogEventsPath = filepath.Join(t.TempDir(), "events.jsonl")
	req.FakeOpencodeCmd = fakeOpencodeCurlChatCompletionsOnce
	req.ExpectedExit = 0
	return nil
}
```