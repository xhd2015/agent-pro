# Scenario

**Feature**: preset `tool-bash` dequeues one bash `tool_call` AgentEvent per HTTP serve

```
empty exchanges[] -> genQueue [tool_call bash]
POST #1 -> tool_calls response, finish_reason tool_calls
```

## Steps

1. Empty config (`exchanges: []`).
2. `--mock-events-preset=tool-bash`.
3. Send one chat completion request.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ConfigJSON = `{"port": 8080, "exchanges": []}`
	req.MockEventsPreset = "tool-bash"
	req.Requests = []string{
		`{"model":"mock-model","messages":[{"role":"user","content":"run-bash-preset"}]}`,
	}
	return nil
}
```