package openai

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared/constant"
	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/agent/llm/llm-mock/mockconfig"
	"github.com/xhd2015/agent-pro/agent/llm/queue"
)

// EncodeChatCompletion maps a breakpoint batch to a Chat Completions response.
// Prefix think text is merged into message content; omitted from tool_call wire.
func EncodeChatCompletion(model string, batch queue.Batch) openai.ChatCompletion {
	ts := time.Now().Unix()
	id := GenChatID()

	switch batch.Breakpoint.Type {
	case types.ActionToolCall:
		tc := toolCallFromEvent(batch.Breakpoint)
		message := openai.ChatCompletionMessage{
			Role: constant.Assistant("assistant"),
			ToolCalls: []openai.ChatCompletionMessageToolCallUnion{
				{
					ID:   tc.ID,
					Type: tc.Type,
					Function: openai.ChatCompletionMessageFunctionToolCallFunction{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				},
			},
		}
		return openai.ChatCompletion{
			ID:      id,
			Created: ts,
			Model:   openai.ChatModel(model),
			Object:  constant.ChatCompletion("chat.completion"),
			Choices: []openai.ChatCompletionChoice{
				{
					Index:        0,
					FinishReason: "tool_calls",
					Message:      message,
				},
			},
			Usage: openai.CompletionUsage{},
		}
	default:
		content := messageContent(batch)
		message := openai.ChatCompletionMessage{
			Role:    constant.Assistant("assistant"),
			Content: content,
		}
		return openai.ChatCompletion{
			ID:      id,
			Created: ts,
			Model:   openai.ChatModel(model),
			Object:  constant.ChatCompletion("chat.completion"),
			Choices: []openai.ChatCompletionChoice{
				{
					Index:        0,
					FinishReason: "stop",
					Message:      message,
				},
			},
			Usage: openai.CompletionUsage{},
		}
	}
}

// EncodeChatCompletionFromExchange maps a legacy prefix exchange to Chat Completions.
func EncodeChatCompletionFromExchange(model string, resp mockconfig.ExchangeResponse) openai.ChatCompletion {
	finishReason := resp.FinishReason
	if finishReason == "" {
		finishReason = "stop"
	}

	message := openai.ChatCompletionMessage{
		Role: constant.Assistant("assistant"),
	}
	if resp.Content != nil {
		message.Content = *resp.Content
	}
	if resp.ToolCalls != nil {
		tcUnions := make([]openai.ChatCompletionMessageToolCallUnion, len(resp.ToolCalls))
		for i, tc := range resp.ToolCalls {
			tcUnions[i] = openai.ChatCompletionMessageToolCallUnion{
				ID:   tc.ID,
				Type: tc.Type,
				Function: openai.ChatCompletionMessageFunctionToolCallFunction{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			}
		}
		message.ToolCalls = tcUnions
	}

	return openai.ChatCompletion{
		ID:      GenChatID(),
		Created: time.Now().Unix(),
		Model:   openai.ChatModel(model),
		Object:  constant.ChatCompletion("chat.completion"),
		Choices: []openai.ChatCompletionChoice{
			{
				Index:        0,
				FinishReason: finishReason,
				Message:      message,
			},
		},
		Usage: openai.CompletionUsage{},
	}
}

// MarshalChatCompletionJSON serializes a chat completion, using null content for tool_calls.
func MarshalChatCompletionJSON(resp mockconfig.ExchangeResponse, completion openai.ChatCompletion) ([]byte, error) {
	data, err := json.Marshal(completion)
	if err != nil {
		return nil, err
	}
	if resp.Content == nil && resp.ToolCalls != nil {
		data = []byte(strings.Replace(string(data), `"content":""`, `"content":null`, 1))
	}
	if completion.Choices[0].FinishReason == "tool_calls" && len(completion.Choices[0].Message.ToolCalls) > 0 {
		data = []byte(strings.Replace(string(data), `"content":""`, `"content":null`, 1))
	}
	return data, nil
}

func messageContent(batch queue.Batch) string {
	think := queue.CollapsedThinkText(batch.PrefixThink)
	msg := batch.Breakpoint.Text
	if think == "" {
		return msg
	}
	if msg == "" {
		return think
	}
	return think + msg
}

// WriteChatStream writes Chat Completions SSE for a breakpoint batch.
func WriteChatStream(w http.ResponseWriter, model string, batch queue.Batch) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming not supported")
	}
	id := GenChatID()
	ts := time.Now().Unix()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	finishReason := "stop"
	var content string
	switch batch.Breakpoint.Type {
	case types.ActionToolCall:
		finishReason = "tool_calls"
		tc := toolCallFromEvent(batch.Breakpoint)
		writeChatStreamChunk(w, flusher, id, ts, model, map[string]any{"role": "assistant"}, "")
		writeChatStreamChunk(w, flusher, id, ts, model, map[string]any{
			"tool_calls": []map[string]any{
				{
					"index": 0,
					"id":    tc.ID,
					"type":  tc.Type,
					"function": map[string]any{
						"name":      tc.Function.Name,
						"arguments": tc.Function.Arguments,
					},
				},
			},
		}, "")
	default:
		content = messageContent(batch)
		writeChatStreamChunk(w, flusher, id, ts, model, map[string]any{"role": "assistant"}, "")
		for i := 0; i < len(content); i += 3 {
			end := i + 3
			if end > len(content) {
				end = len(content)
			}
			writeChatStreamChunk(w, flusher, id, ts, model, map[string]any{"content": content[i:end]}, "")
		}
	}
	writeChatStreamChunk(w, flusher, id, ts, model, map[string]any{}, finishReason)
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
	return nil
}

// WriteChatStreamFromExchange writes Chat Completions SSE for a legacy exchange.
func WriteChatStreamFromExchange(w http.ResponseWriter, model string, exchange mockconfig.ExchangeResponse) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming not supported")
	}
	id := GenChatID()
	ts := time.Now().Unix()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	finishReason := exchange.FinishReason
	if finishReason == "" {
		finishReason = "stop"
	}

	content := ""
	if exchange.Content != nil {
		content = *exchange.Content
	}

	writeChatStreamChunk(w, flusher, id, ts, model, map[string]any{"role": "assistant"}, "")
	for i := 0; i < len(content); i += 3 {
		end := i + 3
		if end > len(content) {
			end = len(content)
		}
		writeChatStreamChunk(w, flusher, id, ts, model, map[string]any{"content": content[i:end]}, "")
	}
	writeChatStreamChunk(w, flusher, id, ts, model, map[string]any{}, finishReason)
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
	return nil
}

func writeChatStreamChunk(w io.Writer, flusher http.Flusher, id string, created int64, model string, delta map[string]any, finishReason string) {
	choice := map[string]any{
		"index": 0,
		"delta": delta,
	}
	if finishReason != "" {
		choice["finish_reason"] = finishReason
	}
	data, _ := json.Marshal(map[string]any{
		"id":      id,
		"object":  "chat.completion.chunk",
		"created": created,
		"model":   model,
		"choices": []map[string]any{choice},
	})
	fmt.Fprintf(w, "data: %s\n\n", string(data))
	flusher.Flush()
}

func toolCallFromEvent(evt types.AgentEvent) mockconfig.ToolCall {
	args, _ := json.Marshal(evt.ToolInput)
	if len(args) == 0 {
		args = []byte("{}")
	}
	id := evt.ID
	if id == "" {
		id = GenFuncCallID()
	}
	return mockconfig.ToolCall{
		ID:   id,
		Type: "function",
		Function: mockconfig.ToolFunction{
			Name:      evt.Tool,
			Arguments: string(args),
		},
	}
}