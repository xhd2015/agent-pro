# Scenario

**Feature**: Responses API streams reasoning item before function_call when think precedes tool_call

```
genQueue [think, tool_call, message]
POST #1 /v1/responses stream -> reasoning (think text) + function_call bash
#1 must NOT emit message breakpoint on wire
```

## Steps

1. `--mock-events-preset=think-tool-message`.
2. `Endpoint` = `/v1/responses`.
3. Send one streaming Responses API request.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Endpoint = "/v1/responses"
	req.MockEventsPreset = "think-tool-message"
	req.Requests = []string{
		`{"model":"mock-model","input":[{"role":"user","content":"responses-reasoning-turn-1"}],"stream":true}`,
	}
	return nil
}
```