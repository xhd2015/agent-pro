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

	// Create the handler
	handler := &mockHandler{
		config:    config,
		exchanges: exchangesList,
		counter:   0,
		requests:  make([]RecordedRequest, 0),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/chat/completions", handler.handleChatCompletions)
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
	config    Config
	exchanges []parsedExchange
	counter   int
	mu        sync.Mutex
	requests  []RecordedRequest
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

	response := map[string]any{
		"requests": requests,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
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
	h.mu.Unlock()

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

func (h *mockHandler) findMatch(messages []openai.ChatCompletionMessageParamUnion) *Exchange {
	// Find last user message
	userMsg := findLastUserMessage(messages)

	for i := range h.exchanges {
		ex := h.exchanges[i].Exchange

		// Check index
		if ex.Request.Index >= 0 {
			if h.counter != ex.Request.Index {
				continue
			}
		}
		// index == -1 means sequential (no counter constraint)

		// Check role
		if ex.Request.Role != "" {
			if userMsg == nil || userMsg.Role != ex.Request.Role {
				continue
			}
		}

		// Check content (substring match)
		if ex.Request.Content != "" {
			if userMsg == nil || !strings.Contains(userMsg.Content, ex.Request.Content) {
				continue
			}
		}

		// Match found
		h.counter++
		return &ex
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

// findLastUserMessage returns the last message with role "user", or the last message as fallback.
func findLastUserMessage(messages []openai.ChatCompletionMessageParamUnion) *messageInfo {
	for i := len(messages) - 1; i >= 0; i-- {
		info := extractMessageInfo(messages[i])
		if info.Role == "user" {
			return info
		}
	}
	// Fallback: return the last message of any role
	if len(messages) > 0 {
		return extractMessageInfo(messages[len(messages)-1])
	}
	return nil
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
