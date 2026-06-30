package agentui

import (
	"fmt"
	"strings"
	"time"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func newAssistantStreamID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}

func emitPhasedAssistantMessage(emit func(types.AgentEvent) error, streamID, text string) error {
	text = strings.TrimSpace(text)
	if streamID == "" {
		streamID = newAssistantStreamID()
	}
	ts := time.Now().UnixMilli()
	if err := emit(types.AgentEvent{
		ID:        streamID,
		Type:      types.ActionMessage,
		Role:      "assistant",
		Phase:     types.PhaseStart,
		Timestamp: ts,
	}); err != nil {
		return err
	}

	chunkSize := 8
	if text == "" {
		if err := emit(types.AgentEvent{
			ID:        streamID,
			Type:      types.ActionMessage,
			Role:      "assistant",
			Phase:     types.PhaseUpdate,
			Timestamp: time.Now().UnixMilli(),
		}); err != nil {
			return err
		}
	} else {
		for end := chunkSize; end < len(text)+chunkSize; end += chunkSize {
			if end > len(text) {
				end = len(text)
			}
			if err := emit(types.AgentEvent{
				ID:        streamID,
				Type:      types.ActionMessage,
				Role:      "assistant",
				Phase:     types.PhaseUpdate,
				Text:      text[:end],
				Timestamp: time.Now().UnixMilli(),
			}); err != nil {
				return err
			}
			if end == len(text) {
				break
			}
		}
	}

	return emit(types.AgentEvent{
		ID:        streamID,
		Type:      types.ActionMessage,
		Role:      "assistant",
		Phase:     types.PhaseEnd,
		Text:      text,
		Timestamp: time.Now().UnixMilli(),
	})
}