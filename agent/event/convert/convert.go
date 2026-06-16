package convert

import (
	"encoding/json"
	"fmt"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/agent/event/codex_types"
	"github.com/xhd2015/agent-pro/agent/event/crush_types"
	"github.com/xhd2015/agent-pro/agent/event/opencode_types"
	"github.com/xhd2015/agent-pro/agent/event/pi_types"
)

func ConvertRawLine(raw []byte, agentRunner string) ([]types.AgentEvent, error) {
	switch agentRunner {
	case "opencode":
		var events []opencode_types.Event
		if err := json.Unmarshal(raw, &events); err != nil {
			return nil, fmt.Errorf("unmarshal opencode events: %w", err)
		}
		return opencode_types.FromOpencode(events, ""), nil
	case "pi":
		var events []pi_types.Event
		if err := json.Unmarshal(raw, &events); err != nil {
			return nil, fmt.Errorf("unmarshal pi events: %w", err)
		}
		return pi_types.FromPi(events), nil
	case "codex":
		var events []codex_types.Event
		if err := json.Unmarshal(raw, &events); err != nil {
			return nil, fmt.Errorf("unmarshal codex events: %w", err)
		}
		return codex_types.FromCodex(events, ""), nil
	case "crush":
		var events []crush_types.Event
		if err := json.Unmarshal(raw, &events); err != nil {
			return nil, fmt.Errorf("unmarshal crush events: %w", err)
		}
		return crush_types.FromCrush(events, ""), nil
	default:
		return nil, fmt.Errorf("unknown agent runner: %s", agentRunner)
	}
}
