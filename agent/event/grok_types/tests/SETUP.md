## Preconditions

- The `grok_types` package must exist under `agent/event/grok_types` with `types.go` and `convert.go`.
- `ToGrok` converts `[]types.AgentEvent` to `[]grok_types.Event`.
- `FromGrok` converts `[]grok_types.Event` to `[]types.AgentEvent`.
- `ToGrok` accepts a `sessionID string` parameter.

## Steps

1. For `Target="to_grok"`: call `grok_types.ToGrok(Events, SessionID)` and marshal the result.
2. For `Target="from_grok"`: call `grok_types.FromGrok(GrokEvents)` and marshal the result.
3. For `Target="roundtrip"`: call `grok_types.ToGrok(Events, SessionID)` then `grok_types.FromGrok(...)` and marshal the result.

```go
import (
	"encoding/json"
	"strings"
	"testing"

	grok_types "github.com/xhd2015/agent-pro/agent/event/grok_types"
	types "github.com/xhd2015/agent-pro/agent/event/types"
)

type Request struct {
	Target     string              // "to_grok", "from_grok", "roundtrip"
	Events     []types.AgentEvent   // input for ToGrok / roundtrip
	GrokEvents []grok_types.Event   // input for FromGrok
	SessionID  string              // session ID for ToGrok
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
	switch req.Target {
	case "to_grok":
		grokEvts := grok_types.ToGrok(req.Events, req.SessionID)
		data, _ := json.Marshal(grokEvts)
		output = string(data)
	case "from_grok":
		agentEvts := grok_types.FromGrok(req.GrokEvents)
		data, _ := json.Marshal(agentEvts)
		output = string(data)
	case "roundtrip":
		grokEvts := grok_types.ToGrok(req.Events, req.SessionID)
		agentEvts := grok_types.FromGrok(grokEvts)
		data, _ := json.Marshal(agentEvts)
		output = string(data)
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
