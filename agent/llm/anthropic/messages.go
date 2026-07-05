package anthropic

import (
	"encoding/json"
	"fmt"
	"net/http"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/agent/llm/openai"
	"github.com/xhd2015/agent-pro/agent/llm/queue"
)

// WriteMessagesNonStream writes an Anthropic Messages API JSON response.
func WriteMessagesNonStream(w http.ResponseWriter, model string, batch queue.Batch) {
	content := buildContentBlocks(batch)
	stopReason := "end_turn"
	for _, block := range content {
		if typ, _ := block["type"].(string); typ == "tool_use" {
			stopReason = "tool_use"
			break
		}
	}

	response := map[string]any{
		"id":          openai.GenMsgID(),
		"type":        "message",
		"role":        "assistant",
		"model":       model,
		"content":     content,
		"stop_reason": stopReason,
		"usage": map[string]any{
			"input_tokens":  0,
			"output_tokens": 0,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func buildContentBlocks(batch queue.Batch) []map[string]any {
	var blocks []map[string]any
	for _, evt := range batch.PrefixThink {
		if evt.Text == "" {
			continue
		}
		blocks = append(blocks, map[string]any{
			"type":     "thinking",
			"thinking": evt.Text,
		})
	}

	switch batch.Breakpoint.Type {
	case types.ActionToolCall:
		input := batch.Breakpoint.ToolInput
		if input == nil {
			input = map[string]any{}
		}
		id := batch.Breakpoint.ID
		if id == "" {
			id = openai.GenFuncCallID()
		}
		blocks = append(blocks, map[string]any{
			"type":  "tool_use",
			"id":    id,
			"name":  batch.Breakpoint.Tool,
			"input": input,
		})
	default:
		text := batch.Breakpoint.Text
		if text != "" {
			blocks = append(blocks, map[string]any{
				"type": "text",
				"text": text,
			})
		}
	}
	return blocks
}

// WriteMessagesStream writes Anthropic SSE events for a breakpoint batch.
func WriteMessagesStream(w http.ResponseWriter, model string, batch queue.Batch) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming not supported")
	}

	msgID := openai.GenMsgID()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	writeSSE(w, flusher, map[string]any{
		"type": "message_start",
		"message": map[string]any{
			"id":    msgID,
			"type":  "message",
			"role":  "assistant",
			"model": model,
			"usage": map[string]any{"input_tokens": 0, "output_tokens": 0},
		},
	})

	blocks := buildContentBlocks(batch)
	for i, block := range blocks {
		writeSSE(w, flusher, map[string]any{
			"type":  "content_block_start",
			"index": i,
			"content_block": block,
		})
		writeSSE(w, flusher, map[string]any{
			"type":  "content_block_stop",
			"index": i,
		})
	}

	stopReason := "end_turn"
	for _, block := range blocks {
		if typ, _ := block["type"].(string); typ == "tool_use" {
			stopReason = "tool_use"
			break
		}
	}

	writeSSE(w, flusher, map[string]any{
		"type": "message_delta",
		"delta": map[string]any{
			"stop_reason":   stopReason,
			"stop_sequence": nil,
		},
		"usage": map[string]any{"output_tokens": 0},
	})
	writeSSE(w, flusher, map[string]any{"type": "message_stop"})
	return nil
}

func writeSSE(w http.ResponseWriter, flusher http.Flusher, obj map[string]any) {
	data, _ := json.Marshal(obj)
	w.Write([]byte("event: "))
	w.Write([]byte(obj["type"].(string)))
	w.Write([]byte("\ndata: "))
	w.Write(data)
	w.Write([]byte("\n\n"))
	flusher.Flush()
}