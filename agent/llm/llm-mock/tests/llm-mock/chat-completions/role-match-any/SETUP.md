# Scenario

**Feature**: empty role matches any

```
POST system role -> empty-role exchange
```

## Steps
1. Write config with an exchange where `request.role` is empty (should match any role).
2. Send a request where the last user message has role "system".
3. The exchange should still match because empty role means match-any.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.ConfigJSON = `{
  "port": 8080,
  "exchanges": [
    {
      "request": {
        "role": "",
        "content": "system message",
        "index": -1
      },
      "response": {
        "content": "matched any role",
        "finish_reason": "stop"
      }
    }
  ]
}`
    req.Requests = []string{
        `{"model":"gpt-4","messages":[{"role":"system","content":"system message"}]}`,
    }
    return nil
}
```
