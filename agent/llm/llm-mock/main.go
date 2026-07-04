package main

import (
	_ "embed"
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
	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/agent/llm/llm-mock/mockconfig"
	"github.com/xhd2015/agent-pro/agent/llm/llm-mock/mockgen"
	"github.com/xhd2015/agent-pro/agent/llm/llm-mock/mockpreset"
	runpkg "github.com/xhd2015/agent-pro/agent/llm/llm-mock/run"
	"github.com/xhd2015/agent-pro/pkgs/fake-agent/events"
	"github.com/xhd2015/skills/install"
)

//go:embed SKILL.md
var skillContent string

const skillName = "llm-mock"

const mainHelp = `
Usage: llm-mock <command> [options]

Commands:
  run       Run grok with a background mock LLM server (llm-mock run --help)
  skill     Show or install the llm-mock skill (llm-mock skill show)

Server mode (default, no subcommand):
  llm-mock [--config FILE] [--mock-events-preset NAME] [--events-file FILE] [--agent-events-file FILE] [--log-http FILE]

  OpenAI-compatible mock HTTP server for /v1/chat/completions, /v1/models, etc.

Options:
  -config string
        Path to JSON config file (or LLM_MOCK_CONFIG / LLM_MOCK_CONFIG_FILE)
  -mock-events-preset string
        Named AgentEvent sequence to seed genQueue after config exchanges, or "list" for catalog
  -events-file string
        Path to append RecordedRequest JSONL (admin/debug)
  -agent-events-file string
        Path to append served AgentEvent JSONL
  -log-http string
        Path to append full HTTP request/response exchange JSONL (must end with .jsonl)

  -h, --help
        Show this overview (server mode: also shows Go flag help)
`



// RecordedRequest stores a received HTTP request for the admin endpoint.
type RecordedRequest struct {
	Index  int            `json:"index"`
	Method string         `json:"method"`
	Path   string         `json:"path"`
	Body   map[string]any `json:"body"`
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 || isHelpArg(args[0]) {
		printMainHelp()
		return
	}
	if len(args) > 0 && args[0] == "skill" {
		if err := handleSkillCommand(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "llm-mock: %v\n", err)
			os.Exit(1)
		}
		return
	}
	if len(args) > 0 && args[0] == "run" {
		if err := handleRunCommand(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "llm-mock: %v\n", err)
			os.Exit(1)
		}
		return
	}

	configPath := flag.String("config", "", "Path to JSON config file")
	mockEventsPreset := flag.String("mock-events-preset", "", "Named AgentEvent sequence for genQueue, or \"list\" for catalog")
	eventsFile := flag.String("events-file", "", "Path to write request events as JSON lines")
	agentEventsFile := flag.String("agent-events-file", "", "Path to write served AgentEvents as JSON lines")
	logHTTPPath := flag.String("log-http", "", "Path to append HTTP exchange JSONL (must end with .jsonl)")
	flag.Parse()

	if *mockEventsPreset == "list" {
		mockpreset.PrintList(os.Stdout)
		return
	}

	if *logHTTPPath != "" && !strings.HasSuffix(*logHTTPPath, ".jsonl") {
		log.Fatalf("--log-http path must end with .jsonl")
	}

	var presetEvents []types.AgentEvent
	if *mockEventsPreset != "" {
		var err error
		presetEvents, err = mockpreset.Resolve(*mockEventsPreset)
		if err != nil {
			log.Fatal(err)
		}
	}

	loaded, err := mockconfig.LoadMerged(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	config := loaded.Config
	exchangesList := loaded.Exchanges
	effectiveIndices := loaded.EffectiveIndices

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

	var agentEventsWriter io.WriteCloser
	if *agentEventsFile != "" {
		f, err := os.OpenFile(*agentEventsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("failed to open agent events file: %v", err)
		}
		defer f.Close()
		agentEventsWriter = f
	}

	var httpLog *httpLogger
	if *logHTTPPath != "" {
		var err error
		httpLog, err = newHTTPLogger(*logHTTPPath)
		if err != nil {
			log.Fatal(err)
		}
		defer httpLog.Close()
	}

	// Create the handler
	handler := &mockHandler{
		config:            config,
		exchanges:         exchangesList,
		effectiveIndices:  effectiveIndices,
		counter:           0,
		genQueue:          append([]types.AgentEvent(nil), presetEvents...),
		requests:          make([]RecordedRequest, 0),
		eventsWriter:      eventsWriter,
		agentEventsWriter: agentEventsWriter,
	}

	mux := http.NewServeMux()
	registerHandler(mux, "/v1/chat/completions", handler.handleChatCompletions, httpLog)
	registerHandler(mux, "/v1/responses", handler.handleResponses, httpLog)
	registerHandler(mux, "/v1/models", handler.handleModels, httpLog)
	registerHandler(mux, "/admin/requests", handler.handleAdminRequests, httpLog)

	// Port fallback: try port, port+1, ..., port+99
	var listener net.Listener
	for offset := 0; offset < 100; offset++ {
		addr := fmt.Sprintf(":%d", config.Port+offset)
		l, err := net.Listen("tcp", addr)
		if err == nil {
			listener = l
			// Print the listening port to stdout (the test harness reads this).
			if tcpAddr, ok := listener.Addr().(*net.TCPAddr); ok {
				fmt.Printf(":%d\n", tcpAddr.Port)
			} else {
				fmt.Println(addr)
			}
			break
		}
	}
	if listener == nil {
		l, err := net.Listen("tcp", ":0")
		if err != nil {
			log.Fatalf("could not bind to any port in range %d-%d: %v", config.Port, config.Port+99, err)
		}
		listener = l
		if tcpAddr, ok := listener.Addr().(*net.TCPAddr); ok {
			fmt.Printf(":%d\n", tcpAddr.Port)
		}
	}

	if err := http.Serve(listener, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func isHelpArg(arg string) bool {
	return arg == "-h" || arg == "--help" || arg == "-help" || arg == "help"
}

func printMainHelp() {
	fmt.Print(strings.TrimPrefix(mainHelp, "\n"))
}

func handleRunCommand(args []string) error {
	args = runpkg.PrependRunFlagsFromEnv(args)
	opts, remain, err := runpkg.ParseRunFlags(args)
	if err != nil {
		return err
	}
	if opts.MockEventsPreset == "list" {
		mockpreset.PrintList(os.Stdout)
		return nil
	}
	if len(remain) < 1 || remain[0] != "grok" {
		return fmt.Errorf("usage: llm-mock run [--mock-events-preset NAME] [--log-events FILE] [--log-http FILE] grok [grok-args...]\n(hint: llm-mock run --help)")
	}
	return runpkg.RunGrok(remain[1:], opts)
}

func handleSkillCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("expected skill show or skill install")
	}
	switch args[0] {
	case "show":
		if len(args) > 1 {
			return fmt.Errorf("unexpected arguments after show")
		}
		fmt.Print(skillContent)
		return nil
	case "install":
		return install.HandleInstall(install.InstallOptions{
			SkillDirName: skillName,
			SkillContent: skillContent,
			Usage:        "llm-mock skill install",
		}, args[1:])
	default:
		return fmt.Errorf("unknown skill sub-command: %s, expected skill show or skill install", args[0])
	}
}

type mockHandler struct {
	config           mockconfig.Config
	exchanges        []mockconfig.ParsedExchange
	effectiveIndices []int
	counter          int
	genQueue         []types.AgentEvent
	genStream        *events.EventStream
	genMu            sync.Mutex
	mu               sync.Mutex
	requests          []RecordedRequest
	eventsWriter      io.Writer
	agentEventsWriter io.Writer
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
	exchange, agentEvents := h.findMatch(req.Messages)
	if exchange == nil {
		writeError(w, "no matching exchange", "no_match", http.StatusBadRequest)
		return
	}
	h.writeAgentEvents(agentEvents)

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
func (h *mockHandler) handleResponsesNonStream(w http.ResponseWriter, model string, exchange *mockconfig.Exchange) {
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
func (h *mockHandler) handleResponsesStream(w http.ResponseWriter, model string, exchange *mockconfig.Exchange) {
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

	seq := 0
	writeRespStreamEvent(w, flusher, &seq, map[string]any{
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

	writeRespStreamEvent(w, flusher, &seq, map[string]any{
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

	writeRespStreamEvent(w, flusher, &seq, map[string]any{
		"type":          "response.content_part.added",
		"item_id":       msgID,
		"output_index":  0,
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
		delta := content[i:end]
		writeRespStreamEvent(w, flusher, &seq, map[string]any{
			"type":          "response.output_text.delta",
			"item_id":       msgID,
			"output_index":  0,
			"content_index": 0,
			"delta":         delta,
		})
	}

	writeRespStreamEvent(w, flusher, &seq, map[string]any{
		"type":          "response.content_part.done",
		"item_id":       msgID,
		"output_index":  0,
		"content_index": 0,
		"part": map[string]any{
			"type":        "output_text",
			"text":        content,
			"annotations": []any{},
		},
	})

	writeRespStreamEvent(w, flusher, &seq, map[string]any{
		"type":         "response.output_item.done",
		"output_index": 0,
		"item": map[string]any{
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
		},
	})

	writeRespStreamEvent(w, flusher, &seq, map[string]any{
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
							"type":        "output_text",
							"text":        content,
							"annotations": []any{},
						},
					},
				},
			},
			"usage": map[string]any{
				"input_tokens":  0,
				"output_tokens": 0,
				"total_tokens":  0,
				"input_tokens_details": map[string]any{
					"cached_tokens": 0,
				},
				"output_tokens_details": map[string]any{
					"reasoning_tokens": 0,
				},
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
	exchange, agentEvents := h.findMatch(req.Messages)
	if exchange == nil {
		writeError(w, "no matching exchange", "no_match", http.StatusBadRequest)
		return
	}
	h.writeAgentEvents(agentEvents)

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

// writeAgentEvents appends served AgentEvent lines when --agent-events-file is set.
func (h *mockHandler) writeAgentEvents(events []types.AgentEvent) {
	if h.agentEventsWriter == nil || len(events) == 0 {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, evt := range events {
		data, _ := json.Marshal(evt)
		fmt.Fprintln(h.agentEventsWriter, string(data))
	}
}

func (h *mockHandler) findMatch(messages []openai.ChatCompletionMessageParamUnion) (*mockconfig.Exchange, []types.AgentEvent) {
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
		return &ex, mockgen.ExchangeResponseToAgentEvents(ex.Response)
	}

	// Prefix exchange expected but request did not match.
	if h.counter < len(h.exchanges) {
		return nil, nil
	}

	// Prefix exhausted: dequeue generated AgentEvent fallback.
	return h.findGeneratedMatch(messages)
}

func (h *mockHandler) findGeneratedMatch(messages []openai.ChatCompletionMessageParamUnion) (*mockconfig.Exchange, []types.AgentEvent) {
	prompt := extractPrompt(messages)

	h.genMu.Lock()
	if len(h.genQueue) > 0 {
		evt := h.genQueue[0]
		h.genQueue = h.genQueue[1:]
		h.genMu.Unlock()
		resp := mockgen.AgentEventToExchangeResponse(sanitizeGeneratedEvent(evt))
		return &mockconfig.Exchange{Response: resp}, []types.AgentEvent{evt}
	}

	if h.genStream == nil {
		seed := mockgen.SeedFromPrompt(prompt)
		h.genStream = events.NewMockEventStream(seed, prompt)
	}
	stream := h.genStream
	h.genMu.Unlock()

	// Do not hold genMu (or handler mu) across probe execution; slow tool probes
	// must not block other HTTP handlers or starve the first think response.
	evt, ok := stream.Next()
	if !ok {
		// Shared stream yields think then message; once exhausted, start a fresh
		// cycle seeded from the latest user prompt for the next turn.
		h.genMu.Lock()
		seed := mockgen.SeedFromPrompt(prompt)
		h.genStream = events.NewMockEventStream(seed, prompt)
		stream = h.genStream
		h.genMu.Unlock()
		evt, ok = stream.Next()
		if !ok {
			return nil, nil
		}
	}
	resp := mockgen.AgentEventToExchangeResponse(sanitizeGeneratedEvent(evt))
	return &mockconfig.Exchange{Response: resp}, []types.AgentEvent{evt}
}

// sanitizeGeneratedEvent normalizes generated events for HTTP JSON responses.
// Think text uses embedded newlines; when tests capture bodies via shell echo,
// JSON \n escapes are interpreted as line breaks and truncate parsed output.
func sanitizeGeneratedEvent(evt types.AgentEvent) types.AgentEvent {
	if evt.Type == types.ActionThink && strings.Contains(evt.Text, "\n") {
		evt.Text = strings.ReplaceAll(evt.Text, "\n", " ")
	}
	return evt
}

func extractPrompt(messages []openai.ChatCompletionMessageParamUnion) string {
	for i := len(messages) - 1; i >= 0; i-- {
		info := extractMessageInfo(messages[i])
		if info == nil || info.Role != "user" || info.Content == "" {
			continue
		}
		return info.Content
	}
	for i := len(messages) - 1; i >= 0; i-- {
		info := extractMessageInfo(messages[i])
		if info != nil && info.Content != "" {
			return info.Content
		}
	}
	return ""
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

func (h *mockHandler) handleNonStream(w http.ResponseWriter, model string, exchange *mockconfig.Exchange) {
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

func (h *mockHandler) handleStream(w http.ResponseWriter, model string, exchange *mockconfig.Exchange) {
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

	writeChatStreamChunk(w, flusher, id, ts, model, map[string]any{"role": "assistant"}, "")

	for i := 0; i < len(content); i += 3 {
		end := i + 3
		if end > len(content) {
			end = len(content)
		}
		chunkContent := content[i:end]
		writeChatStreamChunk(w, flusher, id, ts, model, map[string]any{"content": chunkContent}, "")
	}

	writeChatStreamChunk(w, flusher, id, ts, model, map[string]any{}, finishReason)

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

func writeRespStreamEvent(w io.Writer, flusher http.Flusher, seq *int, obj map[string]any) {
	obj["sequence_number"] = *seq
	*seq++
	writeRespSSE(w, flusher, obj)
}

func writeChatStreamChunk(w io.Writer, flusher http.Flusher, id string, created int64, model string, delta map[string]any, finishReason string) {
	choice := map[string]any{
		"index": 0,
		"delta": delta,
	}
	if finishReason != "" {
		choice["finish_reason"] = finishReason
	}
	writeRespSSE(w, flusher, map[string]any{
		"id":      id,
		"object":  "chat.completion.chunk",
		"created": created,
		"model":   model,
		"choices": []map[string]any{choice},
	})
}
