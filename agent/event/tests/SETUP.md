## Preconditions
- Each leaf provides structured input data in request fields.
- `Run` dispatches based on which fields are set: `Events` → ToCodex/ToOpencode, `Value` → marshal, `Output` → pass-through.

## Steps
1. If `req.Events` is set, call `codex_types.ToCodex` or `opencode_types.ToOpencode` and marshal.
2. If `req.Value` is set, marshal it to JSON.
3. Otherwise, pass through `req.Output`.

```go
import (
	"encoding/json"
	"strings"
	"testing"

	codex_types "github.com/xhd2015/agent-pro/agent/event/codex_types"
	opencode_types "github.com/xhd2015/agent-pro/agent/event/opencode_types"
	types "github.com/xhd2015/agent-pro/agent/event/types"
)

type Request struct {
	Events    []types.AgentEvent
	Target    string // "opencode" → ToOpencode; default → ToCodex
	SessionID string
	Value     any
	Output    string
}

type Response struct {
	Output string
}

func Setup(t *testing.T, req *Request) error {
	_ = assertContains
	_ = assertNotContains
	return nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
	var output string
	if len(req.Events) > 0 {
		if req.Target == "opencode" {
			result := opencode_types.ToOpencode(req.Events, req.SessionID)
			data, _ := json.Marshal(result)
			output = string(data)
		} else {
			result := codex_types.ToCodex(req.Events)
			data, _ := json.Marshal(result)
			output = string(data)
		}
	} else if req.Value != nil {
		data, _ := json.Marshal(req.Value)
		output = string(data)
	} else {
		output = req.Output
	}
	return &Response{Output: output}, nil
}

func assertContains(t *testing.T, got string, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in:\n%s", want, got)
	}
}

func assertNotContains(t *testing.T, got string, want string) {
	t.Helper()
	if strings.Contains(got, want) {
		t.Fatalf("unexpected %q in:\n%s", want, got)
	}
}
```
