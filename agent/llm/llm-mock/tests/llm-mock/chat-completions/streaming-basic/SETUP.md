## Steps
1. Write config with a single exchange matching the request.
2. Send a streaming request (`"stream": true`).
3. The server returns SSE events with chunks and `[DONE]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
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
