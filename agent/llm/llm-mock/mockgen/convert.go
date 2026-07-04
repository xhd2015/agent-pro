package mockgen

import (
	"encoding/json"
	"hash/fnv"
	"os"
	"strconv"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/agent/llm/llm-mock/mockconfig"
)

// SeedFromPrompt returns a deterministic seed from the prompt, with optional
// LLM_MOCK_RANDOM_SEED env override.
func SeedFromPrompt(prompt string) int64 {
	if v := os.Getenv("LLM_MOCK_RANDOM_SEED"); v != "" {
		if seed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return seed
		}
	}
	h := fnv.New64a()
	_, _ = h.Write([]byte(prompt))
	return int64(h.Sum64())
}

// ExchangeResponseToAgentEvents synthesizes AgentEvents from a prefix exchange response.
func ExchangeResponseToAgentEvents(resp mockconfig.ExchangeResponse) []types.AgentEvent {
	var out []types.AgentEvent
	if resp.Content != nil && *resp.Content != "" {
		out = append(out, types.AgentEvent{
			Type: types.ActionMessage,
			Text: *resp.Content,
		})
	}
	for _, tc := range resp.ToolCalls {
		var toolInput map[string]any
		if tc.Function.Arguments != "" {
			_ = json.Unmarshal([]byte(tc.Function.Arguments), &toolInput)
		}
		if toolInput == nil {
			toolInput = map[string]any{}
		}
		out = append(out, types.AgentEvent{
			Type:      types.ActionToolCall,
			ID:        tc.ID,
			Tool:      tc.Function.Name,
			ToolInput: toolInput,
		})
	}
	return out
}

// AgentEventToExchangeResponse converts a generated AgentEvent into the mock
// server's ExchangeResponse shape.
func AgentEventToExchangeResponse(evt types.AgentEvent) mockconfig.ExchangeResponse {
	switch evt.Type {
	case types.ActionThink, types.ActionMessage:
		content := evt.Text
		return mockconfig.ExchangeResponse{
			Content:      &content,
			FinishReason: "stop",
		}
	case types.ActionToolCall:
		args, _ := json.Marshal(evt.ToolInput)
		if len(args) == 0 {
			args = []byte("{}")
		}
		return mockconfig.ExchangeResponse{
			Content: nil,
			ToolCalls: []mockconfig.ToolCall{
				{
					ID:   evt.ID,
					Type: "function",
					Function: mockconfig.ToolFunction{
						Name:      evt.Tool,
						Arguments: string(args),
					},
				},
			},
			FinishReason: "tool_calls",
		}
	default:
		content := evt.Text
		if content == "" {
			content = string(evt.Type)
		}
		return mockconfig.ExchangeResponse{
			Content:      &content,
			FinishReason: "stop",
		}
	}
}