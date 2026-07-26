# Scenario

**Feature**: orchestrator passes `--mock-events-preset=think-message` to mock; fake grok curls twice

```
llm-mock run --mock-events-preset=think-message --log-events f.jsonl grok
orchestrator -> mock genQueue [think, message]
fake grok curl #1 -> think AgentEvent
fake grok curl #2 -> message AgentEvent
```

## Steps

1. No config env (default empty `exchanges[]`).
2. `--mock-events-preset=think-message` and `--log-events` for AgentEvent proof.
3. Fake grok curls mock twice via `GROK_MODELS_BASE_URL`.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

const fakeGrokCurlThinkMessagePreset = `sh -c '
base="${GROK_MODELS_BASE_URL}"
r1=$(curl -sf "$base/chat/completions" -H "Content-Type: application/json" -d "{\"model\":\"mock-model\",\"messages\":[{\"role\":\"user\",\"content\":\"preset-curl-1\"}]}")
r2=$(curl -sf "$base/chat/completions" -H "Content-Type: application/json" -d "{\"model\":\"mock-model\",\"messages\":[{\"role\":\"user\",\"content\":\"preset-curl-2\"}]}")
echo "R1=$r1"
echo "R2=$r2"
'`

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ConfigJSON = ""
	req.ConfigEnv = ""
	req.MockEventsPreset = "think-message"
	req.LogEventsPath = filepath.Join(t.TempDir(), "session.jsonl")
	req.FakeGrokCmd = fakeGrokCurlThinkMessagePreset
	req.ExpectedExit = 0
	return nil
}
```