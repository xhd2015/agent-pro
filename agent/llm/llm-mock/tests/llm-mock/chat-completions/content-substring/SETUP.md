# Scenario

**Feature**: substring content match

```
POST user message -> substring exchange
```

## Steps
1. Write config with exchange where `request.content` is a substring of the expected message.
2. Send a request where the last user message contains that substring.
3. The exchange matches because content is matched as a substring.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.ConfigJSON = `{
  "port": 8080,
  "exchanges": [
    {
      "request": {
        "role": "user",
        "content": "French capital",
        "index": -1
      },
      "response": {
        "content": "Paris is the answer",
        "finish_reason": "stop"
      }
    }
  ]
}`
    req.Requests = []string{
        `{"model":"gpt-4","messages":[{"role":"user","content":"Tell me the French capital please"}]}`,
    }
    return nil
}
```
