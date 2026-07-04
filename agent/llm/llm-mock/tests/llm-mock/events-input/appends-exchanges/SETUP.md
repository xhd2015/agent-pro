# Scenario

**Feature**: events file appends second exchange; second HTTP response comes from events

```
config exchange[0] + events JSONL exchange[1] -> merged server config
POST #1 -> from-config; POST #2 -> from-events
```

## Steps

1. Config JSON with one exchange matching `config-first-prompt`.
2. Events JSONL with one exchange matching `events-second-prompt`.
3. Send two chat completion requests in order.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.ConfigJSON = `{
  "port": 8080,
  "exchanges": [
    {
      "request": {
        "role": "user",
        "content": "config-first-prompt",
        "index": -1
      },
      "response": {
        "content": "from-config",
        "finish_reason": "stop"
      }
    }
  ]
}`
    req.EventsInputJSONL = `{"request":{"role":"user","content":"events-second-prompt","index":-1},"response":{"content":"from-events","finish_reason":"stop"}}` + "\n"
    req.Requests = []string{
        `{"model":"mock-model","messages":[{"role":"user","content":"config-first-prompt"}]}`,
        `{"model":"mock-model","messages":[{"role":"user","content":"events-second-prompt"}]}`,
    }
    return nil
}
```
