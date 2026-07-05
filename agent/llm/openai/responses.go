package openai

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/agent/llm/llm-mock/mockconfig"
	"github.com/xhd2015/agent-pro/agent/llm/queue"
)

func responsesUsage() map[string]any {
	return map[string]any{
		"input_tokens":  0,
		"output_tokens": 0,
		"total_tokens":  0,
	}
}

func responsesCompletedUsage() map[string]any {
	return map[string]any{
		"input_tokens":  0,
		"output_tokens": 0,
		"total_tokens":  0,
		"input_tokens_details": map[string]any{
			"cached_tokens": 0,
		},
		"output_tokens_details": map[string]any{
			"reasoning_tokens": 0,
		},
	}
}

func reasoningItem(id, text string) map[string]any {
	return map[string]any{
		"type":   "reasoning",
		"id":     id,
		"status": "completed",
		"summary": []map[string]any{
			{
				"type": "summary_text",
				"text": text,
			},
		},
	}
}

func functionCallItem(tc mockconfig.ToolCall) map[string]any {
	itemID := tc.ID
	if itemID == "" {
		itemID = GenFuncCallID()
	}
	callID := itemID
	name, arguments := WireFunctionCall(tc)
	return map[string]any{
		"type":      "function_call",
		"id":        itemID,
		"status":    "completed",
		"name":      name,
		"call_id":   callID,
		"arguments": arguments,
	}
}

func messageItem(msgID, content string) map[string]any {
	return map[string]any{
		"type":   "message",
		"id":     msgID,
		"status": "completed",
		"role":   "assistant",
		"content": []map[string]any{
			{
				"type": "output_text",
				"text": content,
			},
		},
	}
}

// BuildResponsesOutput builds non-stream output[] from a breakpoint batch.
func BuildResponsesOutput(batch queue.Batch) []map[string]any {
	thinkText := queue.CollapsedThinkText(batch.PrefixThink)
	switch batch.Breakpoint.Type {
	case types.ActionToolCall:
		var output []map[string]any
		if thinkText != "" {
			output = append(output, reasoningItem(GenReasoningID(), thinkText))
		}
		output = append(output, functionCallItem(toolCallFromEvent(batch.Breakpoint)))
		return output
	default:
		content := messageContent(batch)
		if thinkText != "" && batch.Breakpoint.Text != "" {
			var output []map[string]any
			output = append(output, reasoningItem(GenReasoningID(), thinkText))
			output = append(output, messageItem(GenMsgID(), batch.Breakpoint.Text))
			return output
		}
		return []map[string]any{messageItem(GenMsgID(), content)}
	}
}

// WriteResponsesNonStream writes a completed Responses API JSON body.
func WriteResponsesNonStream(w http.ResponseWriter, model string, batch queue.Batch) {
	respID := GenRespID()
	ts := time.Now().Unix()
	response := map[string]any{
		"id":         respID,
		"object":     "response",
		"created_at": ts,
		"model":      model,
		"status":     "completed",
		"output":     BuildResponsesOutput(batch),
		"usage":      responsesUsage(),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// WriteResponsesNonStreamFromExchange writes legacy prefix exchange as Responses API JSON.
func WriteResponsesNonStreamFromExchange(w http.ResponseWriter, model string, exchange mockconfig.ExchangeResponse) {
	respID := GenRespID()
	ts := time.Now().Unix()
	var output []map[string]any
	if useToolCalls(exchange) {
		for _, tc := range exchange.ToolCalls {
			output = append(output, functionCallItem(tc))
		}
	} else {
		content := ""
		if exchange.Content != nil {
			content = *exchange.Content
		}
		output = []map[string]any{messageItem(GenMsgID(), content)}
	}
	response := map[string]any{
		"id":         respID,
		"object":     "response",
		"created_at": ts,
		"model":      model,
		"status":     "completed",
		"output":     output,
		"usage":      responsesUsage(),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func useToolCalls(resp mockconfig.ExchangeResponse) bool {
	if len(resp.ToolCalls) == 0 {
		return false
	}
	if resp.FinishReason == "tool_calls" {
		return true
	}
	return resp.Content == nil || (resp.Content != nil && *resp.Content == "")
}

// WriteResponsesStream writes Responses API SSE for a breakpoint batch.
func WriteResponsesStream(w http.ResponseWriter, model string, batch queue.Batch) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming not supported")
	}
	respID := GenRespID()
	ts := time.Now().Unix()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	seq := 0
	writeStreamCreated(w, flusher, &seq, respID, model, ts)

	thinkText := queue.CollapsedThinkText(batch.PrefixThink)
	outputIndex := 0
	var output []map[string]any

	if thinkText != "" {
		rsID := GenReasoningID()
		writeReasoningStream(w, flusher, &seq, outputIndex, rsID, thinkText)
		done := reasoningItem(rsID, thinkText)
		output = append(output, done)
		outputIndex++
	}

	switch batch.Breakpoint.Type {
	case types.ActionToolCall:
		tc := toolCallFromEvent(batch.Breakpoint)
		done := writeFunctionCallStream(w, flusher, &seq, outputIndex, tc)
		output = append(output, done)
	default:
		content := batch.Breakpoint.Text
		if thinkText == "" {
			content = messageContent(batch)
		}
		msgID := GenMsgID()
		done := writeMessageStream(w, flusher, &seq, outputIndex, respID, model, ts, msgID, content)
		output = append(output, done)
	}

	writeStreamCompleted(w, flusher, &seq, respID, model, ts, output)
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
	return nil
}

// WriteResponsesStreamFromExchange writes Responses API SSE for a legacy exchange.
func WriteResponsesStreamFromExchange(w http.ResponseWriter, model string, exchange mockconfig.ExchangeResponse) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming not supported")
	}
	respID := GenRespID()
	ts := time.Now().Unix()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	seq := 0
	writeStreamCreated(w, flusher, &seq, respID, model, ts)

	var output []map[string]any
	if useToolCalls(exchange) {
		for i, tc := range exchange.ToolCalls {
			done := writeFunctionCallStream(w, flusher, &seq, i, tc)
			output = append(output, done)
		}
	} else {
		content := ""
		if exchange.Content != nil {
			content = *exchange.Content
		}
		msgID := GenMsgID()
		done := writeMessageStream(w, flusher, &seq, 0, respID, model, ts, msgID, content)
		output = append(output, done)
	}

	writeStreamCompleted(w, flusher, &seq, respID, model, ts, output)
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
	return nil
}

func writeStreamCreated(w io.Writer, flusher http.Flusher, seq *int, respID, model string, ts int64) {
	writeStreamEvent(w, flusher, seq, map[string]any{
		"type": "response.created",
		"response": map[string]any{
			"id":         respID,
			"object":     "response",
			"created_at": ts,
			"model":      model,
			"status":     "in_progress",
			"output":     []any{},
		},
	})
}

func writeStreamCompleted(w io.Writer, flusher http.Flusher, seq *int, respID, model string, ts int64, output []map[string]any) {
	writeStreamEvent(w, flusher, seq, map[string]any{
		"type": "response.completed",
		"response": map[string]any{
			"id":         respID,
			"object":     "response",
			"created_at": ts,
			"model":      model,
			"status":     "completed",
			"output":     output,
			"usage":      responsesCompletedUsage(),
		},
	})
}

func writeReasoningStream(w io.Writer, flusher http.Flusher, seq *int, outputIndex int, rsID, text string) {
	writeStreamEvent(w, flusher, seq, map[string]any{
		"type":         "response.output_item.added",
		"output_index": outputIndex,
		"item": map[string]any{
			"type":   "reasoning",
			"id":     rsID,
			"status": "in_progress",
		},
	})
	writeStreamEvent(w, flusher, seq, map[string]any{
		"type":         "response.output_item.done",
		"output_index": outputIndex,
		"item":         reasoningItem(rsID, text),
	})
}

func writeFunctionCallStream(w io.Writer, flusher http.Flusher, seq *int, outputIndex int, tc mockconfig.ToolCall) map[string]any {
	itemID := tc.ID
	if itemID == "" {
		itemID = GenFuncCallID()
	}
	callID := itemID
	name, arguments := WireFunctionCall(tc)

	writeStreamEvent(w, flusher, seq, map[string]any{
		"type":         "response.output_item.added",
		"output_index": outputIndex,
		"item": map[string]any{
			"type":    "function_call",
			"id":      itemID,
			"status":  "in_progress",
			"name":    name,
			"call_id": callID,
		},
	})

	for i := 0; i < len(arguments); i += 3 {
		end := i + 3
		if end > len(arguments) {
			end = len(arguments)
		}
		writeStreamEvent(w, flusher, seq, map[string]any{
			"type":         "response.function_call_arguments.delta",
			"item_id":      itemID,
			"output_index": outputIndex,
			"delta":        arguments[i:end],
		})
	}

	writeStreamEvent(w, flusher, seq, map[string]any{
		"type":         "response.function_call_arguments.done",
		"item_id":      itemID,
		"output_index": outputIndex,
		"arguments":    arguments,
	})

	doneItem := functionCallItem(tc)
	doneItem["id"] = itemID
	doneItem["call_id"] = callID
	writeStreamEvent(w, flusher, seq, map[string]any{
		"type":         "response.output_item.done",
		"output_index": outputIndex,
		"item":         doneItem,
	})
	return doneItem
}

func writeMessageStream(w io.Writer, flusher http.Flusher, seq *int, outputIndex int, respID, model string, ts int64, msgID, content string) map[string]any {
	writeStreamEvent(w, flusher, seq, map[string]any{
		"type":         "response.output_item.added",
		"output_index": outputIndex,
		"item": map[string]any{
			"type":    "message",
			"id":      msgID,
			"status":  "in_progress",
			"role":    "assistant",
			"content": []any{},
		},
	})

	writeStreamEvent(w, flusher, seq, map[string]any{
		"type":          "response.content_part.added",
		"item_id":       msgID,
		"output_index":  outputIndex,
		"content_index": 0,
		"part": map[string]any{
			"type":        "output_text",
			"text":        "",
			"annotations": []any{},
		},
	})

	for i := 0; i < len(content); i += 3 {
		end := i + 3
		if end > len(content) {
			end = len(content)
		}
		writeStreamEvent(w, flusher, seq, map[string]any{
			"type":          "response.output_text.delta",
			"item_id":       msgID,
			"output_index":  outputIndex,
			"content_index": 0,
			"delta":         content[i:end],
		})
	}

	writeStreamEvent(w, flusher, seq, map[string]any{
		"type":          "response.content_part.done",
		"item_id":       msgID,
		"output_index":  outputIndex,
		"content_index": 0,
		"part": map[string]any{
			"type":        "output_text",
			"text":        content,
			"annotations": []any{},
		},
	})

	doneItem := map[string]any{
		"type":   "message",
		"id":     msgID,
		"status": "completed",
		"role":   "assistant",
		"content": []map[string]any{
			{
				"type":        "output_text",
				"text":        content,
				"annotations": []any{},
			},
		},
	}
	writeStreamEvent(w, flusher, seq, map[string]any{
		"type":         "response.output_item.done",
		"output_index": outputIndex,
		"item":         doneItem,
	})
	return doneItem
}

func writeStreamEvent(w io.Writer, flusher http.Flusher, seq *int, obj map[string]any) {
	obj["sequence_number"] = *seq
	*seq++
	data, err := json.Marshal(obj)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "data: %s\n\n", string(data))
	flusher.Flush()
}