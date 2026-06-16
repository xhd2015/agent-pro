## Steps
1. Write config with an exchange that does NOT match the request.
2. Send a streaming request (`"stream": true`).
3. The server returns HTTP 400 with error JSON (before starting SSE).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.ConfigJSON = `{
  "port": 8080,
  "exchanges": [
    {
      "request": {
        "role": "user",
        "content": "only known message",
        "index": -1
      },
      "response": {
        "content": "matched",
        "finish_reason": "stop"
      }
    }
  ]
}`
    req.Requests = []string{
        `{"model":"gpt-4","messages":[{"role":"user","content":"unknown message"}],"stream":true}`,
    }
    return nil
}
```
