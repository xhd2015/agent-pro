package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/agent/llm/openai"
	"github.com/xhd2015/agent-pro/agent/llm/queue"
	anthropicenc "github.com/xhd2015/agent-pro/agent/llm/anthropic"
)

// WriteChatNonStream writes a non-streaming Chat Completions JSON response.
func WriteChatNonStream(w http.ResponseWriter, model string, serve *serveResult) {
	if serve.fromLegacy {
		completion := openai.EncodeChatCompletionFromExchange(model, serve.legacy.Response)
		data, err := openai.MarshalChatCompletionJSON(serve.legacy.Response, completion)
		if err != nil {
			writeError(w, "failed to encode response", "internal_error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
		return
	}
	completion := openai.EncodeChatCompletion(model, serve.batch)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if serve.batch.Breakpoint.Type == types.ActionToolCall {
		data, _ := json.Marshal(completion)
		w.Write(data)
		return
	}
	json.NewEncoder(w).Encode(completion)
}

// WriteChatStream writes Chat Completions SSE for a breakpoint batch.
func WriteChatStream(w http.ResponseWriter, model string, serve *serveResult) {
	if serve.fromLegacy {
		openai.WriteChatStreamFromExchange(w, model, serve.legacy.Response)
		return
	}
	openai.WriteChatStream(w, model, serve.batch)
}

// WriteMessagesResponse handles both stream and non-stream Anthropic messages.
func WriteMessagesResponse(w http.ResponseWriter, model string, serve *serveResult, stream bool) {
	if serve.fromLegacy {
		writeError(w, "anthropic endpoint requires generated preset queue", "no_match", http.StatusBadRequest)
		return
	}
	if stream {
		anthropicenc.WriteMessagesStream(w, model, serve.batch)
		return
	}
	anthropicenc.WriteMessagesNonStream(w, model, serve.batch)
}

// WriteResponsesNonStream writes a non-streaming Responses API JSON response.
func WriteResponsesNonStream(w http.ResponseWriter, model string, serve *serveResult) {
	if serve.fromLegacy {
		openai.WriteResponsesNonStreamFromExchange(w, model, serve.legacy.Response)
		return
	}
	openai.WriteResponsesNonStream(w, model, serve.batch)
}

// WriteResponsesStream writes streaming Responses API SSE.
func WriteResponsesStream(w http.ResponseWriter, model string, serve *serveResult) {
	if serve.fromLegacy {
		openai.WriteResponsesStreamFromExchange(w, model, serve.legacy.Response)
		return
	}
	openai.WriteResponsesStream(w, model, serve.batch)
}

// --- Alpha NDJSON streaming ---

// WriteAlphaNDJSON streams a serveResult as NDJSON for /alpha/generate.
func WriteAlphaNDJSON(w http.ResponseWriter, model string, serve *serveResult) {
	if serve.fromLegacy {
		WriteAlphaFallbackNDJSON(w, model)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	for _, evt := range serve.batch.PrefixThink {
		if evt.Text != "" {
			fmt.Fprintf(w, "%s\n", encodeAlphaEvent("text-delta", map[string]any{"text": evt.Text}))
		}
	}

	if serve.batch.Breakpoint.Type == types.ActionToolCall {
		tcID := serve.batch.Breakpoint.ID
		if tcID == "" {
			tcID = openai.GenFuncCallID()
		}
		input := serve.batch.Breakpoint.ToolInput
		if input == nil {
			input = map[string]any{}
		}
		fmt.Fprintf(w, "%s\n", encodeAlphaEvent("tool-use", map[string]any{"tool_call_id": tcID, "tool_name": serve.batch.Breakpoint.Tool}))
		inputJSON, _ := json.Marshal(input)
		fmt.Fprintf(w, "%s\n", encodeAlphaEvent("tool-delta", map[string]any{"text": string(inputJSON)}))
	} else {
		fmt.Fprintf(w, "%s\n", encodeAlphaEvent("text-delta", map[string]any{"text": serve.batch.Breakpoint.Text}))
	}

	fmt.Fprintf(w, "%s\n", encodeAlphaEvent("finish", map[string]any{
		"finish_reason": "end_turn",
		"total_usage":   map[string]any{"input_tokens": 0, "output_tokens": 0},
	}))
	flusher.Flush()
}

// WriteAlphaFallbackNDJSON streams a single text block as NDJSON for /alpha/generate fallback.
func WriteAlphaFallbackNDJSON(w http.ResponseWriter, model string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s\n", encodeAlphaEvent("text-delta", map[string]any{"text": "Hello from llm-mock!"}))
	fmt.Fprintf(w, "%s\n", encodeAlphaEvent("finish", map[string]any{
		"finish_reason": "end_turn",
		"total_usage":   map[string]any{"input_tokens": 0, "output_tokens": 0},
	}))
	flusher.Flush()
}

func encodeAlphaEvent(typ string, extra map[string]any) string {
	event := map[string]any{"type": typ}
	for k, v := range extra {
		event[k] = v
	}
	data, _ := json.Marshal(event)
	return string(data)
}

// --- Helpers ---

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
