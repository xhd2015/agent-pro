package openai

import (
	"encoding/json"

	"github.com/xhd2015/agent-pro/agent/llm/llm-mock/mockconfig"
)

// WireFunctionCall maps mock tool_calls to Responses API function_call wire names/arguments.
// Best-effort codex remap: bash→exec_command, command→cmd.
func WireFunctionCall(tc mockconfig.ToolCall) (name, arguments string) {
	name = tc.Function.Name
	arguments = tc.Function.Arguments
	if name != "bash" {
		return name, arguments
	}
	name = "exec_command"
	var args map[string]any
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return name, arguments
	}
	cmd, _ := args["command"].(string)
	if cmd == "" {
		cmd, _ = args["cmd"].(string)
	}
	if cmd == "" {
		return name, arguments
	}
	out, err := json.Marshal(map[string]string{"cmd": cmd})
	if err != nil {
		return name, arguments
	}
	return name, string(out)
}