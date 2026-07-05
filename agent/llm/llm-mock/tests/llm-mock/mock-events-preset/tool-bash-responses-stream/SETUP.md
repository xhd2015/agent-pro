# Scenario

**Feature**: preset `tool-bash` must stream a Responses API `function_call` (not empty message text)

Reproduces codex first-turn silence: `POST /v1/responses` with `stream=true` currently emits only
an empty `output_text` message while agent-events correctly logs `tool_call` bash.

```
empty exchanges[] -> genQueue [tool_call bash]
POST /v1/responses stream=true -> SSE must include function_call name=bash
```

## Steps

1. Empty config (`exchanges: []`).
2. `--mock-events-preset=tool-bash`.
3. Send one streaming Responses API request (codex wire format).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Endpoint = "/v1/responses"
	req.ConfigJSON = `{"port": 8080, "exchanges": []}`
	req.MockEventsPreset = "tool-bash"
	req.Requests = []string{
		`{"model":"mock-model","input":[{"role":"user","content":"hasdfas"}],"stream":true}`,
	}
	return nil
}
```