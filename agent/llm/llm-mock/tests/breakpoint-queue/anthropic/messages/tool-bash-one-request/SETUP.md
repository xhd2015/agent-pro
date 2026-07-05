# Scenario

**Feature**: single tool_call breakpoint encodes to one Anthropic `tool_use` block

```
genQueue [tool_call bash]
POST /v1/messages -> content[] with one tool_use name=bash
```

## Steps

1. `--mock-events-preset=tool-bash`.
2. Send one Anthropic messages request.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.MockEventsPreset = "tool-bash"
	req.Requests = []string{
		`{"model":"mock-model","max_tokens":1024,"messages":[{"role":"user","content":"anthropic-tool-bash"}]}`,
	}
	return nil
}
```