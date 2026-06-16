## Steps
1. Write config with an exchange that specifies `finish_reason: "length"`.
2. Send a matching request.
3. The response includes `finish_reason: "length"`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.ConfigJSON = `{
  "port": 8080,
  "exchanges": [
    {
      "request": {
        "role": "user",
        "content": "long story",
        "index": -1
      },
      "response": {
        "content": "This story was cut short due to length constraints",
        "finish_reason": "length"
      }
    }
  ]
}`
    req.Requests = []string{
        `{"model":"gpt-4","messages":[{"role":"user","content":"long story"}]}`,
    }
    return nil
}
```
