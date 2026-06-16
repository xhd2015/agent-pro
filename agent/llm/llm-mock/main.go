package main

import (
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared/constant"
)

// Config represents the JSON configuration file.
type Config struct {
	Port      int        `json:"port"`
	Exchanges []Exchange `json:"exchanges"`
}

// Exchange represents a request→response mapping.
type Exchange struct {
	Request  ExchangeRequest  `json:"request"`
	Response ExchangeResponse `json:"response"`
}

// ExchangeRequest defines the matching criteria.
type ExchangeRequest struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Index   int    `json:"index"`
}

// ExchangeResponse defines the response to return.
type ExchangeResponse struct {
	Content      *string    `json:"content"`
	ToolCalls    []ToolCall `json:"tool_calls"`
	FinishReason string     `json:"finish_reason"`
}

// ToolCall represents an OpenAI tool call in the config DSL.
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

// ToolFunction represents a function call within a tool call.
type ToolFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// RecordedRequest stores a received HTTP request for the admin endpoint.
type RecordedRequest struct {
	Index  int            `json:"index"`
	Method string         `json:"method"`
	Path   string         `json:"path"`
	Body   map[string]any `json:"body"`
}

func main() {
	configPath := flag.String("config", "", "Path to JSON config file")
	eventsFile := flag.String("events-file", "", "Path to write request events as JSON lines")
	flag.Parse()

	// Load config from --config flag or LLM_MOCK_CONFIG env var
	if *configPath == "" {
		*configPath = os.Getenv("LLM_MOCK_CONFIG")
	}
	if *configPath == "" {
		log.Fatal("no config provided: use --config flag or LLM_MOCK_CONFIG env var")
	}

	configData, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("failed to read config file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(configData, &config); err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}

	// Set default port
	if config.Port == 0 {
		config.Port = 8080
	}

	// We need to detect whether "index" was explicitly set in JSON.
	var rawCfg struct {
		Port      int `json:"port"`
		Exchanges []struct {
			Request  json.RawMessage `json:"request"`
			Response json.RawMessage `json:"response"`
		} `json:"exchanges"`
	}
	if err := json.Unmarshal(configData, &rawCfg); err != nil {
		log.Fatalf("failed to re-parse config: %v", err)
	}

	// Process each exchange to detect explicit index
	var exchangesList []parsedExchange
	for i, re := range rawCfg.Exchanges {
		var reqMap map[string]any
		if err := json.Unmarshal(re.Request, &reqMap); err != nil {
			log.Fatalf("failed to parse exchange %d request: %v", i, err)
		}
		_, hasIndex := reqMap["index"]

		ex := config.Exchanges[i]
		if !hasIndex {
			ex.Request.Index = -1
		}
		exchangesList = append(exchangesList, parsedExchange{Exchange: ex, HasIndex: hasIndex})
	}

	// Pre-compute effective indices: explicit indices stay as-is;
	// index=-1 exchanges receive sequential implicit indices.
	effectiveIndices := make([]int, len(exchangesList))
	nextIdx := 0
	for i := range exchangesList {
		ex := exchangesList[i].Exchange
		if ex.Request.Index >= 0 {
			effectiveIndices[i] = ex.Request.Index
			if ex.Request.Index >= nextIdx {
				nextIdx = ex.Request.Index + 1
			}
		} else {
			effectiveIndices[i] = nextIdx
			nextIdx++
		}
	}

	// Open events file for append if specified
	var eventsWriter io.WriteCloser
	if *eventsFile != "" {
		f, err := os.OpenFile(*eventsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("failed to open events file: %v", err)
		}
		defer f.Close()
		eventsWriter = f
	}

	// Create the handler
	handler := &mockHandler{
		config:           config,
		exchanges:        exchangesList,
		effectiveIndices: effectiveIndices,
		counter:          0,
		requests:         make([]RecordedRequest, 0),
		eventsWriter:     eventsWriter,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/chat/completions", handler.handleChatCompletions)
	mux.HandleFunc("/v1/responses", handler.handleResponses)
	mux.HandleFunc("/v1/models", handler.handleModels)
	mux.HandleFunc("/admin/requests", handler.handleAdminRequests)

	// Port fallback: try port, port+1, ..., port+99
	var listener net.Listener
	for offset := 0; offset < 100; offset++ {
		addr := fmt.Sprintf(":%d", config.Port+offset)
		l, err := net.Listen("tcp", addr)
		if err == nil {
			listener = l
			// Print the listening address to stdout (the test harness reads this)
			fmt.Println(addr)
			break
		}
	}
	if listener == nil {
		log.Fatalf("could not bind to any port in range %d-%d", config.Port, config.Port+99)
	}

	if err := http.Serve(listener, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

type mockHandler struct {
	config           Config
	exchanges        []parsedExchange
	effectiveIndices []int
	counter          int
	mu               sync.Mutex
	requests         []RecordedRequest
	eventsWriter     io.Writer
}

type parsedExchange struct {
	Exchange Exchange
	HasIndex bool
}

func (h *mockHandler) handleModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]any{
		"object": "list",
		"data": []map[string]any{
			{
				"id":       "mock-model",
				"object":   "model",
				"created":  1,
				"owned_by": "llm-mock",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *mockHandler) handleAdminRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.mu.Lock()
	requests := make([]RecordedRequest, len(h.requests))
	copy(requests, h.requests)
	h.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(requests)
}

func (h *mockHandler) handleResponses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse raw body to extract model and input
	var rawBody map[string]any
	if err := json.Unmarshal(body, &rawBody); err != nil {
		writeError(w, "invalid request JSON", "invalid_request", http.StatusBadRequest)
		return
	}

	// Record the request
	h.mu.Lock()
	h.requests = append(h.requests, RecordedRequest{
		Index:  len(h.requests),
		Method: r.Method,
		Path:   r.URL.Path,
		Body:   rawBody,
	})
	rec := h.requests[len(h.requests)-1]
	h.mu.Unlock()
	h.writeEvent(rec)

	// Translate Responses API format to Chat Completions format.
	// The Responses API uses "input" instead of "messages", and content
	// may be an array of content blocks: [{"type":"input_text","text":"..."}].
	model, _ := rawBody["model"].(string)
	input, _ := rawBody["input"]
	messages := convertResponsesInput(input)

	// Reconstruct a chat-completion-style request body for SDK parsing.
	chatBody := map[string]any{
		"model":    model,
		"messages": messages,
	}
	if stream, ok := rawBody["stream"]; ok {
		chatBody["stream"] = stream
	}
	chatJSON, _ := json.Marshal(chatBody)

	var req openai.ChatCompletionNewParams
	if err := json.Unmarshal(chatJSON, &req); err != nil {
		writeError(w, "invalid request JSON", "invalid_request", http.StatusBadRequest)
		return
	}

	// Find matching exchange
	exchange := h.findMatch(req.Messages)
	if exchange == nil {
		writeError(w, "no matching exchange", "no_match", http.StatusBadRequest)
		return
	}

	stream, _ := rawBody["stream"].(bool)

	if stream {
		h.handleResponsesStream(w, model, exchange)
	} else {
		h.handleResponsesNonStream(w, model, exchange)
	}
}

// convertResponsesInput converts OpenAI Responses API input to Chat Completions messages.
// The Responses API uses content arrays like [{"type":"input_text","text":"..."}]
// and wraps messages in an "input" field instead of "messages".
func convertResponsesInput(input any) []map[string]any {
	inputArr, ok := input.([]any)
	if !ok {
		return nil
	}
	var messages []map[string]any
	for _, item := range inputArr {
		msg, ok := item.(map[string]any)
		if !ok {
			continue
		}
		// Convert content array to string if needed
		role, _ := msg["role"].(string)
		content := convertContentToString(msg["content"])
		messages = append(messages, map[string]any{
			"role":    role,
			"content": content,
		})
	}
	return messages
}

// convertContentToString handles both string content and content-block arrays.
func convertContentToString(content any) string {
	switch c := content.(type) {
	case string:
		return c
	case []any:
		var parts []string
		for _, block := range c {
			b, ok := block.(map[string]any)
			if !ok {
				continue
			}
			text, _ := b["text"].(string)
			if text != "" {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, "")
	}
	return ""
}

// handleResponsesNonStream returns an OpenAI Responses API format response.
func (h *mockHandler) handleResponsesNonStream(w http.ResponseWriter, model string, exchange *Exchange) {
	respID := genRespID()
	msgID := genMsgID()
	ts := time.Now().Unix()

	content := ""
	if exchange.Response.Content != nil {
		content = *exchange.Response.Content
	}

	response := map[string]any{
		"id":         respID,
		"object":     "response",
		"created_at": ts,
		"model":      model,
		"status":     "completed",
		"output": []map[string]any{
			{
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
			},
		},
		"usage": map[string]any{
			"input_tokens":  0,
			"output_tokens": 0,
			"total_tokens":  0,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleResponsesStream returns an OpenAI Responses API SSE stream.
func (h *mockHandler) handleResponsesStream(w http.ResponseWriter, model string, exchange *Exchange) {
	respID := genRespID()
	msgID := genMsgID()
	ts := time.Now().Unix()

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, "streaming not supported", "internal_error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	content := ""
	if exchange.Response.Content != nil {
		content = *exchange.Response.Content
	}

	// Event: response.created
	writeRespSSE(w, flusher, map[string]any{
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

	// Event: response.output_item.added
	writeRespSSE(w, flusher, map[string]any{
		"type":         "response.output_item.added",
		"output_index": 0,
		"item": map[string]any{
			"type":    "message",
			"id":      msgID,
			"status":  "in_progress",
			"role":    "assistant",
			"content": []any{},
		},
	})

	// Event: response.content_part.added
	writeRespSSE(w, flusher, map[string]any{
		"type":          "response.content_part.added",
		"item_id":       msgID,
		"output_index":  0,
		"content_index": 0,
		"part": map[string]any{
			"type": "output_text",
			"text": "",
		},
	})

	// Events: response.output_text.delta (split into chunks)
	for i := 0; i < len(content); i += 3 {
		end := i + 3
		if end > len(content) {
			end = len(content)
		}
		delta := content[i:end]
		writeRespSSE(w, flusher, map[string]any{
			"type":          "response.output_text.delta",
			"item_id":       msgID,
			"output_index":  0,
			"content_index": 0,
			"delta":         delta,
		})
	}

	// Event: response.content_part.done
	writeRespSSE(w, flusher, map[string]any{
		"type":          "response.content_part.done",
		"item_id":       msgID,
		"output_index":  0,
		"content_index": 0,
		"part": map[string]any{
			"type": "output_text",
			"text": content,
		},
	})

	// Event: response.output_item.done
	writeRespSSE(w, flusher, map[string]any{
		"type":         "response.output_item.done",
		"output_index": 0,
		"item": map[string]any{
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
		},
	})

	// Event: response.completed
	writeRespSSE(w, flusher, map[string]any{
		"type": "response.completed",
		"response": map[string]any{
			"id":         respID,
			"object":     "response",
			"created_at": ts,
			"model":      model,
			"status":     "completed",
			"output": []map[string]any{
				{
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
				},
			},
			"usage": map[string]any{
				"input_tokens":  0,
				"output_tokens": 0,
				"total_tokens":  0,
			},
		},
	})

	// [DONE] marker
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

func (h *mockHandler) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse into SDK type for full spec compliance
	var req openai.ChatCompletionNewParams
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, "invalid request JSON", "invalid_request", http.StatusBadRequest)
		return
	}

	// Also parse as raw map for stream detection and request recording
	var rawBody map[string]any
	json.Unmarshal(body, &rawBody)

	// Record the request
	h.mu.Lock()
	h.requests = append(h.requests, RecordedRequest{
		Index:  len(h.requests),
		Method: r.Method,
		Path:   r.URL.Path,
		Body:   rawBody,
	})
	rec := h.requests[len(h.requests)-1]
	h.mu.Unlock()
	h.writeEvent(rec)

	// Find matching exchange
	exchange := h.findMatch(req.Messages)
	if exchange == nil {
		writeError(w, "no matching exchange", "no_match", http.StatusBadRequest)
		return
	}

	stream, _ := rawBody["stream"].(bool)
	model := string(req.Model)

	if stream {
		h.handleStream(w, model, exchange)
	} else {
		h.handleNonStream(w, model, exchange)
	}
}

// writeEvent writes a recorded request to the events file if configured.
func (h *mockHandler) writeEvent(rec RecordedRequest) {
	if h.eventsWriter == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	data, _ := json.Marshal(rec)
	fmt.Fprintln(h.eventsWriter, string(data))
}

func (h *mockHandler) findMatch(messages []openai.ChatCompletionMessageParamUnion) *Exchange {
	// If counter has advanced past all configured exchanges, replay the last one.
	// This handles clients that make multiple follow-up requests beyond the
	// configured exchanges (e.g., opencode agent loops).
	var replay *Exchange
	if len(h.exchanges) > 0 && h.counter >= len(h.exchanges) {
		replay = &h.exchanges[len(h.exchanges)-1].Exchange
	}

	for i := range h.exchanges {
		ex := h.exchanges[i].Exchange

		// Check effective index (explicit or implicit sequential)
		if h.counter != h.effectiveIndices[i] {
			continue
		}

		// Check if ANY message matches the exchange request criteria
		matched := false
		for _, msg := range messages {
			info := extractMessageInfo(msg)
			if info == nil {
				continue
			}
			// Check role
			if ex.Request.Role != "" && info.Role != ex.Request.Role {
				continue
			}
			// Check content (substring match)
			if ex.Request.Content != "" && !strings.Contains(info.Content, ex.Request.Content) {
				continue
			}
			matched = true
			break
		}
		if !matched {
			continue
		}

		// Match found
		h.counter++
		return &ex
	}

	// No exact exchange matched. If we've exhausted all configured exchanges,
	// replay the last one for a limited number of extra requests to support
	// multi-turn clients, then stop to prevent infinite loops.
	const maxReplays = 3
	maxCounter := len(h.exchanges) + maxReplays
	if replay != nil && h.counter < maxCounter {
		h.counter++
		return replay
	}
	return nil
}

// messageInfo holds extracted role and content from a ChatCompletionMessageParamUnion.
type messageInfo struct {
	Role    string
	Content string
}

// extractMessageInfo extracts role and content from an SDK message union type.
func extractMessageInfo(msg openai.ChatCompletionMessageParamUnion) *messageInfo {
	info := &messageInfo{}
	if msg.OfUser != nil {
		info.Role = "user"
		if msg.OfUser.Content.OfString.Valid() {
			info.Content = msg.OfUser.Content.OfString.Value
		}
	} else if msg.OfSystem != nil {
		info.Role = "system"
		if msg.OfSystem.Content.OfString.Valid() {
			info.Content = msg.OfSystem.Content.OfString.Value
		}
	} else if msg.OfAssistant != nil {
		info.Role = "assistant"
		if msg.OfAssistant.Content.OfString.Valid() {
			info.Content = msg.OfAssistant.Content.OfString.Value
		}
	} else if msg.OfDeveloper != nil {
		info.Role = "developer"
		if msg.OfDeveloper.Content.OfString.Valid() {
			info.Content = msg.OfDeveloper.Content.OfString.Value
		}
	} else if msg.OfTool != nil {
		info.Role = "tool"
		if msg.OfTool.Content.OfString.Valid() {
			info.Content = msg.OfTool.Content.OfString.Value
		}
	} else if msg.OfFunction != nil {
		info.Role = "function"
		if msg.OfFunction.Content.Valid() {
			info.Content = msg.OfFunction.Content.Value
		}
	}
	return info
}

func (h *mockHandler) handleNonStream(w http.ResponseWriter, model string, exchange *Exchange) {
	id := genID()
	ts := time.Now().Unix()

	finishReason := exchange.Response.FinishReason
	if finishReason == "" {
		finishReason = "stop"
	}

	// Build message using SDK types
	message := openai.ChatCompletionMessage{
		Role: constant.Assistant("assistant"),
	}

	// Handle content (may be nil for tool_calls)
	if exchange.Response.Content != nil {
		message.Content = *exchange.Response.Content
	}

	// Handle tool_calls using SDK types
	if exchange.Response.ToolCalls != nil {
		tcUnions := make([]openai.ChatCompletionMessageToolCallUnion, len(exchange.Response.ToolCalls))
		for i, tc := range exchange.Response.ToolCalls {
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

	// Build full response using SDK ChatCompletion type
	response := openai.ChatCompletion{
		ID:      id,
		Created: ts,
		Model:   openai.ChatModel(model),
		Object:  constant.ChatCompletion("chat.completion"),
		Choices: []openai.ChatCompletionChoice{
			{
				Index:        0,
				FinishReason: finishReason,
				Message:      message,
			},
		},
		Usage: openai.CompletionUsage{
			PromptTokens:     0,
			CompletionTokens: 0,
			TotalTokens:      0,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// For tool_calls responses, OpenAI returns "content": null.
	// The SDK's ChatCompletionMessage uses string, not *string,
	// so we post-process to replace empty-string content with null.
	if exchange.Response.Content == nil && exchange.Response.ToolCalls != nil {
		data, _ := json.Marshal(response)
		data = []byte(strings.Replace(string(data), `"content":""`, `"content":null`, 1))
		w.Write(data)
	} else {
		json.NewEncoder(w).Encode(response)
	}
}

func (h *mockHandler) handleStream(w http.ResponseWriter, model string, exchange *Exchange) {
	id := genID()
	ts := time.Now().Unix()

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, "streaming not supported", "internal_error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	content := ""
	if exchange.Response.Content != nil {
		content = *exchange.Response.Content
	}

	finishReason := exchange.Response.FinishReason
	if finishReason == "" {
		finishReason = "stop"
	}

	// Event 1: role delta — using SDK ChatCompletionChunk type
	chunk := openai.ChatCompletionChunk{
		ID:      id,
		Created: ts,
		Model:   openai.ChatModel(model),
		Object:  constant.ChatCompletionChunk("chat.completion.chunk"),
		Choices: []openai.ChatCompletionChunkChoice{
			{
				Index: 0,
				Delta: openai.ChatCompletionChunkChoiceDelta{
					Role: "assistant",
				},
			},
		},
	}
	writeSSE(w, flusher, chunk)

	// Content events: split into ~3-char chunks
	for i := 0; i < len(content); i += 3 {
		end := i + 3
		if end > len(content) {
			end = len(content)
		}
		chunkContent := content[i:end]
		chunk := openai.ChatCompletionChunk{
			ID:      id,
			Created: ts,
			Model:   openai.ChatModel(model),
			Object:  constant.ChatCompletionChunk("chat.completion.chunk"),
			Choices: []openai.ChatCompletionChunkChoice{
				{
					Index: 0,
					Delta: openai.ChatCompletionChunkChoiceDelta{
						Content: chunkContent,
					},
				},
			},
		}
		writeSSE(w, flusher, chunk)
	}

	// Final event: empty delta, finish_reason
	finalChunk := openai.ChatCompletionChunk{
		ID:      id,
		Created: ts,
		Model:   openai.ChatModel(model),
		Object:  constant.ChatCompletionChunk("chat.completion.chunk"),
		Choices: []openai.ChatCompletionChunkChoice{
			{
				Index:        0,
				Delta:        openai.ChatCompletionChunkChoiceDelta{},
				FinishReason: finishReason,
			},
		},
	}
	writeSSE(w, flusher, finalChunk)

	// [DONE] marker
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

func writeSSE[T any](w io.Writer, flusher http.Flusher, obj T) {
	data, err := json.Marshal(obj)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "data: %s\n\n", string(data))
	flusher.Flush()
}

func writeError(w http.ResponseWriter, message, typ string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]any{
			"message": message,
			"type":    typ,
		},
	})
}

func genID() string {
	// Generate a simple UUID-like string without external dependencies
	b := make([]byte, 16)
	rand.Read(b)
	// Set version 4
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("chatcmpl-%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func genRespID() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("resp_%x%x%x%x%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func genMsgID() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("msg_%x%x%x%x%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func writeRespSSE(w io.Writer, flusher http.Flusher, obj map[string]any) {
	data, err := json.Marshal(obj)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "data: %s\n\n", string(data))
	flusher.Flush()
}
