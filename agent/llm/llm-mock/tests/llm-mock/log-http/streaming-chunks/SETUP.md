# Scenario

**Feature**: streaming SSE logged as chunks array with data: lines

```
POST stream=true -> SSE chunks -> log-http line with response.stream=true
```

## Steps

1. Write config with one exchange for streaming content.
2. Set `LogHTTPFile` to a fresh `.jsonl` path in temp.
3. Send one streaming chat completion request.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.LogHTTPFile = filepath.Join(t.TempDir(), "http-stream.jsonl")
	req.ConfigJSON = `{
  "port": 8080,
  "exchanges": [
    {
      "request": {
        "role": "user",
        "content": "stream test",
        "index": -1
      },
      "response": {
        "content": "Hello streaming world",
        "finish_reason": "stop"
      }
    }
  ]
}`
	req.Requests = []string{
		`{"model":"gpt-4","messages":[{"role":"user","content":"stream test"}],"stream":true}`,
	}
	return nil
}
```