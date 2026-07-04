# Scenario

**Feature**: preset `think-message` dequeues think then message across two HTTP serves

```
empty exchanges[] -> genQueue [think, message]
POST #1 -> think AgentEvent -> HTTP 200
POST #2 -> message AgentEvent -> HTTP 200
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