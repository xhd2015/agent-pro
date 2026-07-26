# Scenario

**Feature**: two sequential exchanges

```
POST x2 -> Paris then follow-up
```

## Steps
1. Write config with two exchanges, both index=-1:
   - exchange[0]: content="one word of French capital", response="Paris"
   - exchange[1]: content="what did I ask", response="you asked for the French capital"
2. Send first request → matches first exchange, returns "Paris".
3. Send second request → counter now 1, matches second exchange, returns "you asked for the French capital".

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
        "content": "one word of French capital",
        "index": -1
      },
      "response": {
        "content": "Paris",
        "finish_reason": "stop"
      }
    },
    {
      "request": {
        "role": "user",
        "content": "what did I ask",
        "index": -1
      },
      "response": {
        "content": "you asked for the French capital",
        "finish_reason": "stop"
      }
    }
  ]
}`
    req.Requests = []string{
        `{"model":"gpt-4","messages":[{"role":"user","content":"one word of French capital"}]}`,
        `{"model":"gpt-4","messages":[{"role":"user","content":"what did I ask"}]}`,
    }
    return nil
}
```
