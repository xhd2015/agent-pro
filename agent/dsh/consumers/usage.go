package consumers

// UsageID is the consumer that accumulates token usage and cost.
const UsageID = "usage"

// UsageState is the session's token totals.
type UsageState struct {
	// Input, Output, and CacheRead are summed over every assistant message.
	Input     int64 `json:"inputTokens,omitempty"`
	Output    int64 `json:"outputTokens,omitempty"`
	CacheRead int64 `json:"cacheReadTokens,omitempty"`
	// Total is the derived sum of the three, so it cannot drift.
	Total int64 `json:"totalTokens,omitempty"`
	// Messages counts assistant messages that carried usage.
	Messages int64 `json:"messages,omitempty"`
	// Steps counts distinct turns, which bounds the session's cost shape.
	Turns int64 `json:"turns,omitempty"`
}

// Usage folds assistant/message usage objects, one per completed LLM request.
func Usage() Consumer {
	return &Typed[UsageState]{
		Name:  UsageID,
		Ver:   1,
		Watch: []string{"assistant/message", "turn/start"},
		Apply: func(state *UsageState, ev Event, _ Point) error {
			fields := decodeObject(ev.Data)
			if fields == nil {
				return nil
			}
			if ev.Type == "turn/start" {
				state.Turns++
				return nil
			}
			usage := objectField(fields, "usage")
			if usage == nil {
				return nil
			}
			input := integer(usage, "inputTokens")
			output := integer(usage, "outputTokens")
			cacheRead := integer(usage, "cacheReadTokens")
			state.Input += input
			state.Output += output
			state.CacheRead += cacheRead
			state.Total += input + output + cacheRead
			state.Messages++
			return nil
		},
	}
}

// Usage returns the folded totals.
func (s Snapshot) Usage() UsageState {
	var state UsageState
	_ = s.Value(UsageID, &state)
	return state
}
