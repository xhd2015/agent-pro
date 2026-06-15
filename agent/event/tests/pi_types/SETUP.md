## Preconditions

- The `pi_types` package must exist under `agent/event/pi_types` with `types.go` and `convert.go`.
- `types.AgentEvent` has a `Phase EventPhase` field.
- `ToPi` converts `[]types.AgentEvent` to `[]pi_types.Event`.
- `FromPi` converts `[]pi_types.Event` to `[]types.AgentEvent`.

## Steps

1. For `Target="wire"`: unmarshal `JSONInput` into a `pi_types.Event`, then marshal back to JSON.
2. For `Target="to_pi"`: call `pi_types.ToPi(Events)` and marshal the result.
3. For `Target="from_pi"`: call `pi_types.FromPi(PiEvents)` and marshal the result.
4. Default: pass through `Output`.

```go
import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	pi_types "github.com/xhd2015/agent-pro/agent/event/pi_types"
	types "github.com/xhd2015/agent-pro/agent/event/types"
)

type Request struct {
	Target    string           // "wire", "to_pi", "from_pi"
	JSONInput string           // raw JSON for wire format parsing
	Events    []types.AgentEvent // input for ToPi
	PiEvents  []pi_types.Event   // input for FromPi
	Output    string           // passthrough
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
	case "wire":
		var evt pi_types.Event
		if err := json.Unmarshal([]byte(req.JSONInput), &evt); err != nil {
			return nil, fmt.Errorf("unmarshal error: %w", err)
		}
		data, _ := json.Marshal(evt)
		output = string(data)
	case "to_pi":
		piEvts := pi_types.ToPi(req.Events)
		data, _ := json.Marshal(piEvts)
		output = string(data)
	case "from_pi":
		agentEvts := pi_types.FromPi(req.PiEvents)
		data, _ := json.Marshal(agentEvts)
		output = string(data)
	case "roundtrip":
		piEvts := pi_types.ToPi(req.Events)
		agentEvts := pi_types.FromPi(piEvts)
		data, _ := json.Marshal(agentEvts)
		output = string(data)
	default:
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
