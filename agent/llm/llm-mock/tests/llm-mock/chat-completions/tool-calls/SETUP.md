# Scenario

**Feature**: tool_calls response

```
POST -> tool_calls, null content
```

## Steps
1. Write config with an exchange that includes `tool_calls` in the response and `finish_reason: "tool_calls"`.
2. Send a matching request.
3. The response includes a `tool_calls` array and null content.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.ConfigJSON = `{
  "port": 8080,
  "exchanges": [
    {
      "request": {
        "role": "user",
        "content": "weather in Paris",
        "index": -1
      },
      "response": {
        "content": null,
        "tool_calls": [
          {
            "id": "call_1",
            "type": "function",
            "function": {
              "name": "get_weather",
              "arguments": "{\"city\":\"Paris\"}"
            }
          }
        ],
        "finish_reason": "tool_calls"
      }
    }
  ]
}`
    req.Requests = []string{
        `{"model":"gpt-4","messages":[{"role":"user","content":"weather in Paris"}]}`,
    }
    return nil
}
```
