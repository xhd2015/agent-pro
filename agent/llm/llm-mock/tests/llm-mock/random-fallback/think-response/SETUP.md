# Scenario

**Feature**: generated `ActionThink` event maps to chat completion with non-empty `content`

```
prefix exhausted -> GenerateEvents -> ActionThink (first event)
ActionThink -> choices[0].message.content = thought text, finish_reason stop
```

## Steps

1. Config JSON with `exchanges: []` so first request hits generator immediately.
2. Send one chat completion request.
3. First generated event is always `ActionThink` per `events.GenerateEvents`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ConfigJSON = `{
  "port": 8080,
  "exchanges": []
}`
	req.Requests = []string{
		`{"model":"mock-model","messages":[{"role":"user","content":"think-response-prompt for analysis"}]}`,
	}
	return nil
}
```