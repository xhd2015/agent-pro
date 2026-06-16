## Steps
1. Write config with an exchange matching the pi integration scenario (role-match-any works for pi since pi may send messages with different roles).
2. Send a chat completion request (simulating what pi sends: model gpt-4, user message about capital of France).
3. Verify the response contains the expected answer.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.ConfigJSON = `{
  "port": 8080,
  "exchanges": [
    {
      "request": {
        "content": "capital of France",
        "index": -1
      },
      "response": {
        "content": "Paris is the capital of France.",
        "finish_reason": "stop"
      }
    }
  ]
}`
    req.Requests = []string{
        `{"model":"gpt-4","messages":[{"role":"user","content":"What is the capital of France?"}]}`,
    }
    return nil
}
```
