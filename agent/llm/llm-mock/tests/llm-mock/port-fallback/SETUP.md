## Steps
1. Write config with port 8080.
2. Block port 8080 before starting the server.
3. The server should fall back to port 8081.
4. Verify the server is listening on port 8081.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.ConfigJSON = `{
  "port": 8080,
  "exchanges": [
    {
      "request": {
        "role": "user",
        "content": "fallback test",
        "index": -1
      },
      "response": {
        "content": "port fallback works",
        "finish_reason": "stop"
      }
    }
  ]
}`
    req.Port = 8080
    req.BlockPort = 8080
    req.Requests = []string{
        `{"model":"gpt-4","messages":[{"role":"user","content":"fallback test"}]}`,
    }
    return nil
}
```
