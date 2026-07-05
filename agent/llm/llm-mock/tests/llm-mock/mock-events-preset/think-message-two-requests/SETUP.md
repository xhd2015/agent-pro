# Scenario

**Feature**: preset `think-message` dequeues think+message as one breakpoint on first HTTP serve

```
empty exchanges[] -> genQueue [think, message]
POST #1 -> merged think+message content (one breakpoint)
POST #2 -> genStream fallback
```

## Steps

1. Empty config (`exchanges: []`).
2. `--mock-events-preset=think-message`.
3. Send two chat completion requests in order.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ConfigJSON = `{"port": 8080, "exchanges": []}`
	req.MockEventsPreset = "think-message"
	req.Requests = []string{
		`{"model":"mock-model","messages":[{"role":"user","content":"preset-turn-1"}]}`,
		`{"model":"mock-model","messages":[{"role":"user","content":"preset-turn-2"}]}`,
	}
	return nil
}
```