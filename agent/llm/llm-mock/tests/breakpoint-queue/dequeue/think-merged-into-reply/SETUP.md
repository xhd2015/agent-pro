# Scenario

**Feature**: `[think, message]` single HTTP — think text merged into chat reply content

```
genQueue [think, message]
POST #1 chat -> message.content contains think text + message text
agent-events: think + message on same serve
```

## Steps

1. `--mock-events-preset=think-message`.
2. Send one chat completion request.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.MockEventsPreset = "think-message"
	req.Requests = []string{
		`{"model":"mock-model","messages":[{"role":"user","content":"merge-think-reply"}]}`,
	}
	return nil
}
```