# Scenario

**Feature**: non-streaming POST logged with request path and response body

```
POST /v1/chat/completions -> 200 JSON body -> 1 log-http line (stream=false)
```

## Steps

1. Write config with one exchange returning "Paris".
2. Set `LogHTTPFile` to a fresh `.jsonl` path in temp.
3. Send one non-streaming chat completion request.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.LogHTTPFile = filepath.Join(t.TempDir(), "http-debug.jsonl")
	req.ConfigJSON = `{
  "port": 8080,
  "exchanges": [
    {
      "request": {
        "role": "user",
        "content": "one word of French capital",
        "index": 0
      },
      "response": {
        "content": "Paris",
        "finish_reason": "stop"
      }
    }
  ]
}`
	req.Requests = []string{
		`{"model":"gpt-4","messages":[{"role":"user","content":"one word of French capital"}]}`,
	}
	return nil
}
```