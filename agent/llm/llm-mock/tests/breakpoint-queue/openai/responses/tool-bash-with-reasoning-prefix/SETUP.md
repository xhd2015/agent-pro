# Scenario

**Feature**: Responses encoder emits reasoning item before function_call for [think, tool_call]

```
genQueue [think, tool_call, message] — #1 consumes think+tool only
POST /v1/responses stream -> reasoning + function_call (bash remapped)
```

## Steps

1. `--mock-events-preset=think-tool-message`.
2. Send one streaming Responses API request.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.MockEventsPreset = "think-tool-message"
	req.Requests = []string{
		`{"model":"mock-model","input":[{"role":"user","content":"responses-reasoning-encode"}],"stream":true}`,
	}
	return nil
}
```