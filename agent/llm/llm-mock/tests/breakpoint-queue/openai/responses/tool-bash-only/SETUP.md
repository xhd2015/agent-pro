# Scenario

**Feature**: single tool_call breakpoint on Responses API — function_call only, no reasoning item

```
genQueue [tool_call bash]
POST /v1/responses stream -> function_call bash (no leading reasoning)
```

## Steps

1. `--mock-events-preset=tool-bash`.
2. Send one streaming Responses API request.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.MockEventsPreset = "tool-bash"
	req.Requests = []string{
		`{"model":"mock-model","input":[{"role":"user","content":"responses-tool-only"}],"stream":true}`,
	}
	return nil
}
```