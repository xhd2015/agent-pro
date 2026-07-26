# Scenario

**Feature**: `--log-events` records think then message AgentEvents from random fallback

```
no mock config -> random fallback on each curl
fake grok -> two curls mock (Hello) simulating grok think+message HTTP split
mock server -> log think on curl #1, message on curl #2 -> session.jsonl
```

## Steps

1. No config env vars (default empty `exchanges[]`).
2. Fake grok performs two identical curls via `GROK_MODELS_BASE_URL` (think then message).
3. `--log-events` points at a fresh `.jsonl` file in temp.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

const fakeGrokCurlTwiceNoConfig = `sh -c '
base="${GROK_MODELS_BASE_URL}"
r1=$(curl -sf "$base/chat/completions" -H "Content-Type: application/json" -d "{\"model\":\"mock-model\",\"messages\":[{\"role\":\"user\",\"content\":\"Hello\"}]}")
r2=$(curl -sf "$base/chat/completions" -H "Content-Type: application/json" -d "{\"model\":\"mock-model\",\"messages\":[{\"role\":\"user\",\"content\":\"Hello\"}]}")
echo "R1=$r1"
echo "R2=$r2"
'`

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ConfigJSON = ""
	req.ConfigEnv = ""
	req.FakeGrokCmd = fakeGrokCurlTwiceNoConfig
	req.LogEventsPath = filepath.Join(t.TempDir(), "session-random-fallback.jsonl")
	return nil
}
```