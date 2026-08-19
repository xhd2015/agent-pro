// Package server provides an in-process llm-mock HTTP server.
// Both the standalone llm-mock binary and the run orchestrators use this package
// so that no sibling binary dependency exists.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/openai/openai-go/v3"
	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/agent/llm/llm-mock/mockconfig"
	"github.com/xhd2015/agent-pro/agent/llm/llm-mock/mockgen"
	"github.com/xhd2015/agent-pro/agent/llm/llm-mock/mockpreset"
	"github.com/xhd2015/agent-pro/agent/llm/queue"
	"github.com/xhd2015/agent-pro/pkgs/fake-agent/events"
)

// Options configures the mock server.
type Options struct {
	ConfigPath       string
	MockEventsPreset string
	MockEventsFile   string
	EventsFile       string
	AgentEventsFile  string
	LogHTTPPath      string
	PresetEvents     []types.AgentEvent
	Config           *mockconfig.Loaded
}

// RecordedRequest stores a received HTTP request.
type RecordedRequest struct {
	Index  int            `json:"index"`
	Method string         `json:"method"`
	Path   string         `json:"path"`
	Body   map[string]any `json:"body"`
}

// Server wraps the mock HTTP server and handler.
type Server struct {
	Handler  *Handler
	listener net.Listener
	HTTPAddr string
	mux      *http.ServeMux
	httpLog  *httpLogger
}

// Handler is the mock request handler.
type Handler struct {
	Config            mockconfig.Config
	Exchanges         []mockconfig.ParsedExchange
	EffectiveIndices  []int
	counter           int
	GenQueue          []types.AgentEvent
	genStream         *events.EventStream
	genMu             sync.Mutex
	mu                sync.Mutex
	Requests          []RecordedRequest
	eventsWriter      io.Writer
	agentEventsWriter io.Writer
}

// Start creates and starts a mock server on an available port.
func Start(ctx context.Context, opts Options) (*Server, error) {
	loaded := opts.Config
	var err error
	if loaded == nil {
		loaded, err = mockconfig.LoadMerged(opts.ConfigPath)
		if err != nil {
			return nil, err
		}
	}

	config := loaded.Config
	exchangesList := loaded.Exchanges
	effectiveIndices := loaded.EffectiveIndices

	presetEvents := opts.PresetEvents
	if opts.MockEventsPreset != "" && len(presetEvents) == 0 {
		var err error
		presetEvents, err = mockpreset.Resolve(opts.MockEventsPreset)
		if err != nil {
			return nil, err
		}
	}
	if strings.TrimSpace(opts.MockEventsFile) != "" {
		extra, err := mockpreset.LoadJSONL(opts.MockEventsFile)
		if err != nil {
			return nil, err
		}
		presetEvents = append(presetEvents, extra...)
	}

	var eventsWriter io.WriteCloser
	if opts.EventsFile != "" {
		f, err := os.OpenFile(opts.EventsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open events file: %w", err)
		}
		eventsWriter = f
	}

	var agentEventsWriter io.WriteCloser
	if opts.AgentEventsFile != "" {
		f, err := os.OpenFile(opts.AgentEventsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open agent events file: %w", err)
		}
		agentEventsWriter = f
	}

	var httpLog *httpLogger
	if opts.LogHTTPPath != "" {
		var err error
		httpLog, err = newHTTPLogger(opts.LogHTTPPath)
		if err != nil {
			return nil, err
		}
	}

	handler := &Handler{
		Config:            config,
		Exchanges:         exchangesList,
		EffectiveIndices:  effectiveIndices,
		GenQueue:          append([]types.AgentEvent(nil), presetEvents...),
		Requests:          make([]RecordedRequest, 0),
		eventsWriter:      eventsWriter,
		agentEventsWriter: agentEventsWriter,
	}

	mux := http.NewServeMux()
	registerHandler(mux, "/v1/chat/completions", handler.handleChatCompletions, httpLog)
	registerHandler(mux, "/v1/responses", handler.handleResponses, httpLog)
	registerHandler(mux, "/v1/messages", handler.handleMessages, httpLog)
	registerHandler(mux, "/v1/models", handler.handleModels, httpLog)
	registerHandler(mux, "/admin/requests", handler.handleAdminRequests, httpLog)
	registerHandler(mux, "/alpha/generate", handler.handleAlphaGenerate, httpLog)
	registerHandler(mux, "/alpha/whoami", handler.handleAlphaWhoami, httpLog)
	registerHandler(mux, "/alpha/lifecycle-events", handler.handleAlphaLifecycleEvents, httpLog)
	registerHandler(mux, "/alpha/fingerprint/record", handler.handleAlphaFingerprintRecord, httpLog)

	var listener net.Listener
	for offset := 0; offset < 100; offset++ {
		addr := fmt.Sprintf(":%d", config.Port+offset)
		l, err := net.Listen("tcp", addr)
		if err == nil {
			listener = l
			break
		}
	}
	if listener == nil {
		l, err := net.Listen("tcp", ":0")
		if err != nil {
			return nil, fmt.Errorf("could not bind to any port: %w", err)
		}
		listener = l
	}

	srv := &Server{
		Handler:  handler,
		listener: listener,
		mux:      mux,
		httpLog:  httpLog,
	}

	if tcpAddr, ok := listener.Addr().(*net.TCPAddr); ok {
		srv.HTTPAddr = fmt.Sprintf("127.0.0.1:%d", tcpAddr.Port)
	}

	go srv.serve(ctx)
	return srv, nil
}

func (s *Server) serve(ctx context.Context) {
	httpSrv := &http.Server{Handler: s.mux}
	go func() {
		<-ctx.Done()
		// One-shot clients (fake curl once) may leave trailing preset messages
		// in GenQueue after a tool_call breakpoint; flush them to --agent-events-file.
		s.Handler.flushRemainingAgentEvents()
		httpSrv.Close()
	}()
	httpSrv.Serve(s.listener)
}

// Port returns the listening port number.
func (s *Server) Port() int {
	if tcpAddr, ok := s.listener.Addr().(*net.TCPAddr); ok {
		return tcpAddr.Port
	}
	return 0
}

// Close shuts down the server.
func (s *Server) Close() {
	if s.Handler != nil {
		s.Handler.flushRemainingAgentEvents()
	}
	if s.httpLog != nil {
		s.httpLog.Close()
	}
}

// flushRemainingAgentEvents appends any unserved GenQueue events to the agent-events log.
func (h *Handler) flushRemainingAgentEvents() {
	if h == nil {
		return
	}
	h.genMu.Lock()
	remaining := h.GenQueue
	h.GenQueue = nil
	h.genMu.Unlock()
	h.writeAgentEvents(remaining)
}

// --- Alpha request types ---

type alphaGenerateRequest struct {
	Config   map[string]any      `json:"config"`
	Memory   string              `json:"memory"`
	Taste    any                 `json:"taste"`
	Skills   string              `json:"skills"`
	Params   alphaGenerateParams `json:"params"`
	ThreadID string              `json:"threadId"`
}

type alphaGenerateParams struct {
	Tools       []map[string]any `json:"tools"`
	Stream      bool             `json:"stream"`
	MaxTokens   int              `json:"max_tokens"`
	Temperature float64          `json:"temperature"`
	Messages    []alphaMessage   `json:"messages"`
	Model       string           `json:"model"`
}

type alphaMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

// --- Handler methods ---

func (h *Handler) handleModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	response := map[string]any{
		"object": "list",
		"data": []map[string]any{
			{"id": "mock-model", "object": "model", "created": 1, "owned_by": "llm-mock"},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) handleAdminRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	h.mu.Lock()
	requests := make([]RecordedRequest, len(h.Requests))
	copy(requests, h.Requests)
	h.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(requests)
}

func (h *Handler) handleResponses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	var rawBody map[string]any
	if err := json.Unmarshal(body, &rawBody); err != nil {
		writeError(w, "invalid request JSON", "invalid_request", http.StatusBadRequest)
		return
	}

	h.recordRequest(r.Method, r.URL.Path, rawBody)

	model, _ := rawBody["model"].(string)
	input, _ := rawBody["input"]
	messages := convertResponsesInput(input)

	chatBody := map[string]any{"model": model, "messages": messages}
	if stream, ok := rawBody["stream"]; ok {
		chatBody["stream"] = stream
	}
	chatJSON, _ := json.Marshal(chatBody)

	var req openai.ChatCompletionNewParams
	if err := json.Unmarshal(chatJSON, &req); err != nil {
		writeError(w, "invalid request JSON", "invalid_request", http.StatusBadRequest)
		return
	}

	serveResult, ok := h.findMatch(req.Messages)
	if !ok {
		writeError(w, "no matching exchange", "no_match", http.StatusBadRequest)
		return
	}
	queue.SleepFor(serveResult.agentEvents)
	h.writeAgentEvents(serveResult.agentEvents)

	stream, _ := rawBody["stream"].(bool)
	if stream {
		WriteResponsesStream(w, model, serveResult)
		return
	}
	WriteResponsesNonStream(w, model, serveResult)
}

func (h *Handler) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	var req openai.ChatCompletionNewParams
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, "invalid request JSON", "invalid_request", http.StatusBadRequest)
		return
	}

	var rawBody map[string]any
	json.Unmarshal(body, &rawBody)

	h.recordRequest(r.Method, r.URL.Path, rawBody)

	serveResult, ok := h.findMatch(req.Messages)
	if !ok {
		writeError(w, "no matching exchange", "no_match", http.StatusBadRequest)
		return
	}
	queue.SleepFor(serveResult.agentEvents)
	h.writeAgentEvents(serveResult.agentEvents)

	stream, _ := rawBody["stream"].(bool)
	model := string(req.Model)

	if stream {
		WriteChatStream(w, model, serveResult)
		return
	}
	WriteChatNonStream(w, model, serveResult)
}

func (h *Handler) handleMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	var rawBody map[string]any
	if err := json.Unmarshal(body, &rawBody); err != nil {
		writeError(w, "invalid request JSON", "invalid_request", http.StatusBadRequest)
		return
	}

	h.recordRequest(r.Method, r.URL.Path, rawBody)

	model, _ := rawBody["model"].(string)
	messages := convertAnthropicMessages(rawBody["messages"])
	chatBody := map[string]any{"model": model, "messages": messages}
	chatJSON, _ := json.Marshal(chatBody)
	var req openai.ChatCompletionNewParams
	if err := json.Unmarshal(chatJSON, &req); err != nil {
		writeError(w, "invalid request JSON", "invalid_request", http.StatusBadRequest)
		return
	}

	serveResult, ok := h.findMatch(req.Messages)
	if !ok {
		writeError(w, "no matching exchange", "no_match", http.StatusBadRequest)
		return
	}
	queue.SleepFor(serveResult.agentEvents)
	h.writeAgentEvents(serveResult.agentEvents)

	stream, _ := rawBody["stream"].(bool)
	WriteMessagesResponse(w, model, serveResult, stream)
}

func (h *Handler) handleAlphaGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	var req alphaGenerateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, "invalid request JSON", "invalid_request", http.StatusBadRequest)
		return
	}

	h.recordRequest(r.Method, r.URL.Path, map[string]any{"model": req.Params.Model, "messages": len(req.Params.Messages)})

	model := req.Params.Model
	if model == "" {
		model = "claude-sonnet-5"
	}

	var chatMsgs []openai.ChatCompletionMessageParamUnion
	for _, msg := range req.Params.Messages {
		text := extractAlphaContentText(msg.Content)
		switch msg.Role {
		case "user":
			chatMsgs = append(chatMsgs, openai.ChatCompletionMessageParamUnion{
				OfUser: &openai.ChatCompletionUserMessageParam{
					Content: openai.ChatCompletionUserMessageParamContentUnion{OfString: openai.String(text)},
				},
			})
		case "assistant":
			chatMsgs = append(chatMsgs, openai.ChatCompletionMessageParamUnion{
				OfAssistant: &openai.ChatCompletionAssistantMessageParam{
					Content: openai.ChatCompletionAssistantMessageParamContentUnion{OfString: openai.String(text)},
				},
			})
		}
	}

	serveResult, ok := h.findMatch(chatMsgs)
	if !ok {
		WriteAlphaFallbackNDJSON(w, model)
		return
	}
	queue.SleepFor(serveResult.agentEvents)
	h.writeAgentEvents(serveResult.agentEvents)
	WriteAlphaNDJSON(w, model, serveResult)
}

func (h *Handler) handleAlphaWhoami(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"user": map[string]any{
			"id": "mock-user-id", "name": "mock-user", "email": "mock@example.com", "userName": "mock-user",
		},
		"org": nil,
	})
}

func (h *Handler) handleAlphaLifecycleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
}

func (h *Handler) handleAlphaFingerprintRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
}

func (h *Handler) recordRequest(method, path string, body map[string]any) {
	h.mu.Lock()
	h.Requests = append(h.Requests, RecordedRequest{Index: len(h.Requests), Method: method, Path: path, Body: body})
	h.mu.Unlock()
}

func (h *Handler) writeAgentEvents(events []types.AgentEvent) {
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

// peekTrailingMessages returns ActionMessage events still in queue without removing them.
func peekTrailingMessages(queue []types.AgentEvent) []types.AgentEvent {
	var msgs []types.AgentEvent
	for _, evt := range queue {
		if evt.Type == types.ActionMessage {
			msgs = append(msgs, evt)
		}
	}
	return msgs
}

// --- Exchange matching (moved from main.go mockHandler) ---

type serveResult struct {
	legacy      *mockconfig.Exchange
	batch       queue.Batch
	agentEvents []types.AgentEvent
	fromLegacy  bool
}

func (h *Handler) findMatch(messages []openai.ChatCompletionMessageParamUnion) (*serveResult, bool) {
	for i := range h.Exchanges {
		ex := h.Exchanges[i].Exchange
		if h.counter != h.EffectiveIndices[i] {
			continue
		}
		matched := false
		for _, msg := range messages {
			info := extractMessageInfo(msg)
			if info == nil {
				continue
			}
			if ex.Request.Role != "" && info.Role != ex.Request.Role {
				continue
			}
			if ex.Request.Content != "" && !strings.Contains(info.Content, ex.Request.Content) {
				continue
			}
			matched = true
			break
		}
		if !matched {
			continue
		}
		h.counter++
		return &serveResult{legacy: &ex, agentEvents: mockgen.ExchangeResponseToAgentEvents(ex.Response), fromLegacy: true}, true
	}
	if h.counter < len(h.Exchanges) {
		return nil, false
	}
	return h.findGeneratedMatch(messages)
}

func (h *Handler) findGeneratedMatch(messages []openai.ChatCompletionMessageParamUnion) (*serveResult, bool) {
	prompt := extractPrompt(messages)

	h.genMu.Lock()
	if batch, ok := queue.DequeueToBreakpoint(&h.GenQueue); ok {
		trailing := peekTrailingMessages(h.GenQueue)
		h.genMu.Unlock()
		batch = sanitizeBatch(batch)
		events := queue.ConsumedEvents(batch)
		if batch.Breakpoint.Type == types.ActionToolCall {
			// Include trailing preset messages in the agent-events log for one-shot
			// clients (fake curl once) without dequeuing them from GenQueue.
			events = append(events, trailing...)
		}
		return &serveResult{batch: batch, agentEvents: events}, true
	}

	if h.genStream == nil {
		seed := mockgen.SeedFromPrompt(prompt)
		h.genStream = events.NewMockEventStream(seed, prompt)
	}
	stream := h.genStream
	h.genMu.Unlock()

	var prefixThink []types.AgentEvent
	for {
		evt, ok := stream.Next()
		if !ok {
			h.genMu.Lock()
			seed := mockgen.SeedFromPrompt(prompt)
			h.genStream = events.NewMockEventStream(seed, prompt)
			stream = h.genStream
			h.genMu.Unlock()
			evt, ok = stream.Next()
			if !ok {
				return nil, false
			}
		}
		evt = sanitizeGeneratedEvent(evt)
		if evt.Type == types.ActionThink {
			prefixThink = append(prefixThink, evt)
			continue
		}
		if evt.Type == types.ActionToolCall || evt.Type == types.ActionMessage {
			batch := queue.Batch{PrefixThink: prefixThink, Breakpoint: evt}
			return &serveResult{batch: batch, agentEvents: queue.ConsumedEvents(batch)}, true
		}
	}
}

func sanitizeBatch(batch queue.Batch) queue.Batch {
	for i := range batch.PrefixThink {
		batch.PrefixThink[i] = sanitizeGeneratedEvent(batch.PrefixThink[i])
	}
	batch.Breakpoint = sanitizeGeneratedEvent(batch.Breakpoint)
	return batch
}

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

type messageInfo struct {
	Role    string
	Content string
}

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

func convertAnthropicMessages(input any) []map[string]any {
	arr, ok := input.([]any)
	if !ok {
		return nil
	}
	var messages []map[string]any
	for _, item := range arr {
		msg, ok := item.(map[string]any)
		if !ok {
			continue
		}
		role, _ := msg["role"].(string)
		content := convertContentToString(msg["content"])
		messages = append(messages, map[string]any{"role": role, "content": content})
	}
	return messages
}

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
		role, _ := msg["role"].(string)
		content := convertContentToString(msg["content"])
		messages = append(messages, map[string]any{"role": role, "content": content})
	}
	return messages
}

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

func writeError(w http.ResponseWriter, message, typ string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"message": message, "type": typ}})
}

// extractAlphaContentText extracts text from an alpha message content field.
// Content can be a plain string (print mode) or an array of content blocks (interactive mode).
func extractAlphaContentText(raw json.RawMessage) string {
	// Try string first (print mode)
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	// Try array of content blocks (interactive mode)
	var blocks []map[string]any
	if err := json.Unmarshal(raw, &blocks); err != nil {
		return ""
	}
	var parts []string
	for _, block := range blocks {
		text, _ := block["text"].(string)
		if text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, "")
}
