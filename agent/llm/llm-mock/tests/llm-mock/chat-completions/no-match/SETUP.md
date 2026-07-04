# Scenario

**Feature**: no match HTTP 400

```
POST unmatched -> HTTP 400
```

## Steps
1. Write config with one exchange that does NOT match the request content.
2. Send a request with content that has no matching exchange.
3. The server returns HTTP 400 with error JSON.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.ConfigJSON = `{
  "port": 8080,
  "exchanges": [
    {
      "request": {
        "role": "user",
        "content": "only this message",
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
        `{"model":"gpt-4","messages":[{"role":"user","content":"completely different message"}]}`,
    }
    return nil
}
```
