## Steps
1. Write config with two exchanges using explicit indices 0 and 1.
2. Send first request → matches by counter 0 == exchange index 0.
3. Send second request → matches by counter 1 == exchange index 1.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.ConfigJSON = `{
  "port": 8080,
  "exchanges": [
    {
      "request": {
        "role": "user",
        "content": "first question",
        "index": 0
      },
      "response": {
        "content": "first answer",
        "finish_reason": "stop"
      }
    },
    {
      "request": {
        "role": "user",
        "content": "second question",
        "index": 1
      },
      "response": {
        "content": "second answer",
        "finish_reason": "stop"
      }
    }
  ]
}`
    req.Requests = []string{
        `{"model":"gpt-4","messages":[{"role":"user","content":"first question"}]}`,
        `{"model":"gpt-4","messages":[{"role":"user","content":"second question"}]}`,
    }
    return nil
}
```
