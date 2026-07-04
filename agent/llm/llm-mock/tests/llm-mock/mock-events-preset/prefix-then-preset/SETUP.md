# Scenario

**Feature**: one config prefix exchange consumed; second HTTP serve dequeues preset `simple` message

```
exchange[0] -> POST #1 returns from-prefix
genQueue [message] -> POST #2 -> preset message (not genStream random)
```

## Steps

1. Config JSON with exactly one prefix exchange (`index: -1`).
2. `--mock-events-preset=simple` (one `message` event).
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
        "content": "prefix-only-prompt",
        "index": -1
      },
      "response": {
        "content": "from-prefix",
        "finish_reason": "stop"
      }
    }
  ]
}`
	req.MockEventsPreset = "simple"
	req.Requests = []string{
		`{"model":"mock-model","messages":[{"role":"user","content":"prefix-only-prompt"}]}`,
		`{"model":"mock-model","messages":[{"role":"user","content":"after-prefix-prompt"}]}`,
	}
	return nil
}
```