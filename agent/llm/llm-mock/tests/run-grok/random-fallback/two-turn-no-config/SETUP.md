# Scenario

**Feature**: `llm-mock run grok` no-config session; second user turn must not error with no_match

```
no mock config -> random fallback
fake grok simulates grok HTTP pattern:
  curl #1 user:Hello (turn 1 think)
  curl #2 user:Hello (turn 1 message)
  curl #3 multi-turn + user:"what's wrong with me?" (turn 2)
```

Reproduces user report: first "Hello" succeeds, then "what's wrong with me?" fails
with `no_match: no matching exchange` on `/v1/chat/completions`.

## Steps

1. No config env vars (default empty `exchanges[]`).
2. Fake grok curls mock three times via `GROK_MODELS_BASE_URL`.
3. Assert all curls return HTTP 200 bodies without `no_match`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
const fakeGrokCurlTwoTurnNoConfig = `sh -c '
base="${GROK_MODELS_BASE_URL}"
r1=$(curl -sf "$base/chat/completions" -H "Content-Type: application/json" -d "{\"model\":\"mock-model\",\"messages\":[{\"role\":\"user\",\"content\":\"Hello\"}]}")
r2=$(curl -sf "$base/chat/completions" -H "Content-Type: application/json" -d "{\"model\":\"mock-model\",\"messages\":[{\"role\":\"user\",\"content\":\"Hello\"}]}")
r3=$(curl -sf "$base/chat/completions" -H "Content-Type: application/json" -d "{\"model\":\"mock-model\",\"messages\":[{\"role\":\"user\",\"content\":\"Hello\"},{\"role\":\"assistant\",\"content\":\"Here is the turn-1 reply.\"},{\"role\":\"user\",\"content\":\"what'\''s wrong with me?\"}]}")
echo "R1=$r1"
echo "R2=$r2"
echo "R3=$r3"
'`

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ConfigJSON = ""
	req.ConfigEnv = ""
	req.FakeGrokCmd = fakeGrokCurlTwoTurnNoConfig
	req.ExpectedExit = 0
	return nil
}
```