# Scenario

**Feature**: second Responses API request after tool breakpoint yields message only

```
genQueue [think, tool_call, message]
POST #1 responses -> think+tool consumed
POST #2 responses -> message breakpoint only
```

## Steps

1. `--mock-events-preset=think-tool-message`.
2. `Endpoint` = `/v1/responses`.
3. Send two Responses API requests (stream on #1 optional; non-stream on #2).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Endpoint = "/v1/responses"
	req.MockEventsPreset = "think-tool-message"
	req.Requests = []string{
		`{"model":"mock-model","input":[{"role":"user","content":"responses-after-tool-1"}],"stream":true}`,
		`{"model":"mock-model","input":[{"role":"user","content":"responses-after-tool-2"}],"stream":false}`,
	}
	return nil
}
```