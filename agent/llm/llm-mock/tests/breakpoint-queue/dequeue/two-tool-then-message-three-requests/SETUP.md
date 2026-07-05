# Scenario

**Feature**: two consecutive tool_call breakpoints then message — one tool per HTTP response

```
genQueue [tool_call bash, tool_call read, message]
POST #1 -> tool_calls bash
POST #2 -> tool_calls read
POST #3 -> message content
```

## Preconditions

- Implementer adds `two-tool-message` preset to `mockpreset` with sequence:
  `tool_call` bash (`echo preset-bash`), `tool_call` read (`preset-read-target.txt`), `message` (`preset:message:two-tool-message`).

## Steps

1. `--mock-events-preset=two-tool-message`.
2. Send three chat completion requests in order.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.MockEventsPreset = "two-tool-message"
	req.Requests = []string{
		`{"model":"mock-model","messages":[{"role":"user","content":"two-tool-turn-1"}]}`,
		`{"model":"mock-model","messages":[{"role":"user","content":"two-tool-turn-2"}]}`,
		`{"model":"mock-model","messages":[{"role":"user","content":"two-tool-turn-3"}]}`,
	}
	return nil
}
```