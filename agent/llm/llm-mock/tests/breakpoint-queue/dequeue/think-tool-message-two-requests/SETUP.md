# Scenario

**Feature**: preset `think-tool-message` dequeues think→tool on #1, message on #2 (chat)

```
genQueue [think, tool_call, message]
POST #1 chat -> tool_calls (think not in wire)
POST #2 chat -> message content
agent-events: think+tool_call on #1, message on #2
```

## Steps

1. `--mock-events-preset=think-tool-message`.
2. Send two chat completion requests in order.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.MockEventsPreset = "think-tool-message"
	req.Requests = []string{
		`{"model":"mock-model","messages":[{"role":"user","content":"bp-turn-1"}]}`,
		`{"model":"mock-model","messages":[{"role":"user","content":"bp-turn-2"}]}`,
	}
	return nil
}
```