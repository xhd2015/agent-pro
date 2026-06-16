## Steps
1. Write config with an exchange matching the opencode integration scenario.
2. Send a chat completion request (simulating what opencode sends: model gpt-4, user message about capital of France).
3. Verify the response contains the expected answer.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.ConfigJSON = `{
  "port": 8080,
  "exchanges": [
    {
      "request": {
        "role": "user",
        "content": "capital of France",
        "index": -1
      },
      "response": {
        "content": "The capital of France is Paris.",
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
