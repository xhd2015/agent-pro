# Scenario

**Feature**: one prefix exchange consumed; second request uses random generator → 200

```
prefix exchange[0] -> POST #1 returns from-prefix
prefix exhausted -> POST #2 -> GenerateEvents -> HTTP 200 (generated, not 400)
```

## Steps

1. Config JSON with exactly one prefix exchange (`index: -1`).
2. Send two chat completion requests in order.
3. First response matches prefix; second is generated fallback.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ConfigJSON = `{
  "port": 8080,
  "exchanges": [
    {
      "request": {
        "role": "user",
        "content": "prefix-only-prompt",
        "index": -1
      },
      "response": {
        "content": "from-prefix",
        "finish_reason": "stop"
      }
    }
  ]
}`
	req.Requests = []string{
		`{"model":"mock-model","messages":[{"role":"user","content":"prefix-only-prompt"}]}`,
		`{"model":"mock-model","messages":[{"role":"user","content":"overflow-to-random-prompt"}]}`,
	}
	return nil
}
```