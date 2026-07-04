# Scenario

**Feature**: exact match index=0 -> Paris

```
POST -> exact role+content -> Paris
```

## Steps
1. Write config with a single exchange: index=0, role="user", content="one word of French capital".
2. Send a chat request with a user message matching that content.
3. The server matches by index=0, returning the configured response.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
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
