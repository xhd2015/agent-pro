# Scenario

**Feature**: OpenAI Responses API encoder (`POST /v1/responses`)

```
breakpoint slice -> responses.go encode -> output[] (reasoning, function_call, output_text)
```

## Steps

1. Default `Endpoint` = `/v1/responses`.
2. Leaves use streaming or non-streaming Responses requests.
3. `Assert` checks reasoning preamble, function_call wire, and codex tool remap.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Endpoint = "/v1/responses"
	return nil
}
```