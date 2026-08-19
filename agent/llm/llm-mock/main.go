package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/openai/openai-go/v3"
	types "github.com/xhd2015/agent-pro/agent/event/types"
	anthropicenc "github.com/xhd2015/agent-pro/agent/llm/anthropic"
	"github.com/xhd2015/agent-pro/agent/llm/llm-mock/mockconfig"
	"github.com/xhd2015/agent-pro/agent/llm/llm-mock/mockgen"
	"github.com/xhd2015/agent-pro/agent/llm/llm-mock/mockpreset"
	runpkg "github.com/xhd2015/agent-pro/agent/llm/llm-mock/run"
	openaienc "github.com/xhd2015/agent-pro/agent/llm/openai"
	"github.com/xhd2015/agent-pro/agent/llm/queue"
	"github.com/xhd2015/agent-pro/pkgs/fake-agent/events"
	lessflags "github.com/xhd2015/less-flags"
	"github.com/xhd2015/skills/install"
)

//go:embed SKILL.md
var skillContent string

const skillName = "llm-mock"

const mainHelp = `
Usage: llm-mock <command> [options]

Commands:
  run       Run grok, codex, or opencode with a background mock LLM server (llm-mock run --help)
  skill     Show or install the llm-mock skill (llm-mock skill show)

Server mode (default, no subcommand):
  llm-mock [--config FILE] [--mock-events-preset NAME] [--mock-events-file FILE] [--events-file FILE] [--agent-events-file FILE] [--log-http FILE]

  OpenAI-compatible mock HTTP server for /v1/chat/completions, /v1/models, etc.

Options:
  -config string
        Path to JSON config file (or LLM_MOCK_CONFIG / LLM_MOCK_CONFIG_FILE)
  -mock-events-preset string
        Named AgentEvent sequence to seed genQueue after config exchanges, or "list" for catalog
  -mock-events-file string
        AgentEvent JSONL appended to genQueue (delay_ms / type=sleep honored before HTTP write)
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
	mockEventsFile := flag.String("mock-events-file", "", "AgentEvent JSONL appended to genQueue")
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
	if strings.TrimSpace(*mockEventsFile) != "" {
		extra, err := mockpreset.LoadJSONL(*mockEventsFile)
		if err != nil {
			log.Fatal(err)
		}
		presetEvents = append(presetEvents, extra...)
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
	registerHandler(mux, "/v1/messages", handler.handleMessages, httpLog)
	registerHandler(mux, "/v1/models", handler.handleModels, httpLog)
	registerHandler(mux, "/admin/requests", handler.handleAdminRequests, httpLog)
	// Alpha endpoints for Command Code sandbox mode
	registerHandler(mux, "/alpha/generate", handler.handleAlphaGenerate, httpLog)
	registerHandler(mux, "/alpha/whoami", handler.handleAlphaWhoami, httpLog)
	registerHandler(mux, "/alpha/lifecycle-events", handler.handleAlphaLifecycleEvents, httpLog)
	registerHandler(mux, "/alpha/fingerprint/record", handler.handleAlphaFingerprintRecord, httpLog)

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

// alphaGenerateRequest mirrors the JSON shape Command Code sends to /alpha/generate.
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

// handleAlphaWhoami returns a minimal auth response so cmd passes the auth check.
func (h *mockHandler) handleAlphaWhoami(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"user": map[string]any{
			"id":       "mock-user-id",
			"name":     "mock-user",
			"email":    "mock@example.com",
			"userName": "mock-user",
		},
		"org": nil,
	})
}

// extractAlphaContentText extracts text from an alpha message content field.
// Content can be a plain string (print mode) or an array of content blocks (interactive mode).
func extractAlphaContentText(raw json.RawMessage) string {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
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
} // handleAlphaLifecycleEvents accepts lifecycle events (no-op).
func (h *mockHandler) handleAlphaLifecycleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
}

// handleAlphaFingerprintRecord accepts fingerprint submissions (no-op).
func (h *mockHandler) handleAlphaFingerprintRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
}

// handleAlphaGenerate handles the Command Code /alpha/generate endpoint.
// Streams newline-delimited JSON (NDJSON) with cc-format event types.
func (h *mockHandler) handleAlphaGenerate(w http.ResponseWriter, r *http.Request) {
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

	var req alphaGenerateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, "invalid request JSON", "invalid_request", http.StatusBadRequest)
		return
	}

	h.mu.Lock()
	h.requests = append(h.requests, RecordedRequest{
		Index:  len(h.requests),
		Method: r.Method,
		Path:   r.URL.Path,
		Body:   map[string]any{"model": req.Params.Model, "messages": len(req.Params.Messages)},
	})
	h.mu.Unlock()

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
					Content: openai.ChatCompletionUserMessageParamContentUnion{
						OfString: openai.String(text),
					},
				},
			})
		case "assistant":
			chatMsgs = append(chatMsgs, openai.ChatCompletionMessageParamUnion{
				OfAssistant: &openai.ChatCompletionAssistantMessageParam{
					Content: openai.ChatCompletionAssistantMessageParamContentUnion{
						OfString: openai.String(text),
					},
				},
			})
		}
	}

	serve, ok := h.matchAndHold(chatMsgs)
	if !ok {
		writeAlphaFallbackNDJSON(w, model, "Hello from llm-mock!")
		return
	}
	h.writeAgentEvents(serve.agentEvents)

	if serve.fromLegacy {
		text := ""
		if serve.legacy.Response.Content != nil {
			text = *serve.legacy.Response.Content
		}
		writeAlphaFallbackNDJSON(w, model, text)
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

	// Emit think events
	for _, evt := range serve.batch.PrefixThink {
		if evt.Text != "" {
			fmt.Fprintf(w, "%s\n", encodeAlphaEvent("text-delta", map[string]any{"text": evt.Text}))
		}
	}

	if serve.batch.Breakpoint.Type == types.ActionToolCall {
		tcID := serve.batch.Breakpoint.ID
		if tcID == "" {
			tcID = openaienc.GenFuncCallID()
		}
		input := serve.batch.Breakpoint.ToolInput
		if input == nil {
			input = map[string]any{}
		}
		fmt.Fprintf(w, "%s\n", encodeAlphaEvent("tool-use", map[string]any{
			"tool_call_id": tcID,
			"tool_name":    serve.batch.Breakpoint.Tool,
		}))
		inputJSON, _ := json.Marshal(input)
		fmt.Fprintf(w, "%s\n", encodeAlphaEvent("tool-delta", map[string]any{
			"text": string(inputJSON),
		}))
	} else {
		text := serve.batch.Breakpoint.Text
		fmt.Fprintf(w, "%s\n", encodeAlphaEvent("text-delta", map[string]any{"text": text}))
	}

	fmt.Fprintf(w, "%s\n", encodeAlphaEvent("finish", map[string]any{
		"finish_reason": "end_turn",
		"total_usage":   map[string]any{"input_tokens": 0, "output_tokens": 0},
	}))
	flusher.Flush()
}

func writeAlphaFallbackNDJSON(w http.ResponseWriter, model, text string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s\n", encodeAlphaEvent("text-delta", map[string]any{"text": text}))
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

func handleRunCommand(args []string) error {
	args = runpkg.PrependRunFlagsFromEnv(args)
	opts, remain, err := runpkg.ParseRunFlags(args)
	if errors.Is(err, lessflags.ErrHelp) {
		return nil
	}
	if err != nil {
		return err
	}
	if opts.MockEventsPreset == "list" {
		mockpreset.PrintList(os.Stdout)
		return nil
	}
	if len(remain) < 1 {
		return fmt.Errorf("usage: llm-mock run [--mock-events-preset NAME] [--log-events FILE] [--log-http FILE] (grok|codex|opencode) [agent-args...]\n(hint: llm-mock run --help)")
	}
	switch remain[0] {
	case "grok":
		return runpkg.RunGrok(remain[1:], opts)
	case "codex":
		return runpkg.RunCodex(remain[1:], runpkg.RunCodexOptions{
			MockEventsPreset: opts.MockEventsPreset,
			MockEventsFile:   opts.MockEventsFile,
			LogEventsPath:    opts.LogEventsPath,
			LogHTTPPath:      opts.LogHTTPPath,
		})
	case "opencode":
		return runpkg.RunOpencode(remain[1:], runpkg.RunOpencodeOptions{
			MockEventsPreset: opts.MockEventsPreset,
			MockEventsFile:   opts.MockEventsFile,
			LogEventsPath:    opts.LogEventsPath,
			LogHTTPPath:      opts.LogHTTPPath,
		})
	case "commandcode":
		return runpkg.RunCommandCode(remain[1:], runpkg.RunCommandCodeOptions{
			MockEventsPreset: opts.MockEventsPreset,
			MockEventsFile:   opts.MockEventsFile,
			LogEventsPath:    opts.LogEventsPath,
			LogHTTPPath:      opts.LogHTTPPath,
		})
	default:
		return fmt.Errorf("usage: llm-mock run [--mock-events-preset NAME] [--log-events FILE] [--log-http FILE] (grok|codex|opencode|commandcode) [agent-args...]\n(hint: llm-mock run --help)")
	}
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
	config            mockconfig.Config
	exchanges         []mockconfig.ParsedExchange
	effectiveIndices  []int
	counter           int
	genQueue          []types.AgentEvent
	genStream         *events.EventStream
	genMu             sync.Mutex
	mu                sync.Mutex
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

	serve, ok := h.matchAndHold(req.Messages)
	if !ok {
		writeError(w, "no matching exchange", "no_match", http.StatusBadRequest)
		return
	}
	h.writeAgentEvents(serve.agentEvents)

	stream, _ := rawBody["stream"].(bool)
	if stream {
		if serve.fromLegacy {
			if err := openaienc.WriteResponsesStreamFromExchange(w, model, serve.legacy.Response); err != nil {
				writeError(w, err.Error(), "internal_error", http.StatusInternalServerError)
			}
			return
		}
		if err := openaienc.WriteResponsesStream(w, model, serve.batch); err != nil {
			writeError(w, err.Error(), "internal_error", http.StatusInternalServerError)
		}
		return
	}
	if serve.fromLegacy {
		openaienc.WriteResponsesNonStreamFromExchange(w, model, serve.legacy.Response)
		return
	}
	openaienc.WriteResponsesNonStream(w, model, serve.batch)
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

	serve, ok := h.matchAndHold(req.Messages)
	if !ok {
		writeError(w, "no matching exchange", "no_match", http.StatusBadRequest)
		return
	}
	h.writeAgentEvents(serve.agentEvents)

	stream, _ := rawBody["stream"].(bool)
	model := string(req.Model)

	if stream {
		if serve.fromLegacy {
			if err := openaienc.WriteChatStreamFromExchange(w, model, serve.legacy.Response); err != nil {
				writeError(w, err.Error(), "internal_error", http.StatusInternalServerError)
			}
			return
		}
		if err := openaienc.WriteChatStream(w, model, serve.batch); err != nil {
			writeError(w, err.Error(), "internal_error", http.StatusInternalServerError)
		}
		return
	}
	if serve.fromLegacy {
		completion := openaienc.EncodeChatCompletionFromExchange(model, serve.legacy.Response)
		data, err := openaienc.MarshalChatCompletionJSON(serve.legacy.Response, completion)
		if err != nil {
			writeError(w, "failed to encode response", "internal_error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
		return
	}
	completion := openaienc.EncodeChatCompletion(model, serve.batch)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if serve.batch.Breakpoint.Type == types.ActionToolCall {
		data, _ := json.Marshal(completion)
		data = []byte(strings.Replace(string(data), `"content":""`, `"content":null`, 1))
		w.Write(data)
		return
	}
	json.NewEncoder(w).Encode(completion)
}

func (h *mockHandler) handleMessages(w http.ResponseWriter, r *http.Request) {
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

	var rawBody map[string]any
	if err := json.Unmarshal(body, &rawBody); err != nil {
		writeError(w, "invalid request JSON", "invalid_request", http.StatusBadRequest)
		return
	}

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

	model, _ := rawBody["model"].(string)
	messages := convertAnthropicMessages(rawBody["messages"])
	chatBody := map[string]any{"model": model, "messages": messages}
	chatJSON, _ := json.Marshal(chatBody)
	var req openai.ChatCompletionNewParams
	if err := json.Unmarshal(chatJSON, &req); err != nil {
		writeError(w, "invalid request JSON", "invalid_request", http.StatusBadRequest)
		return
	}

	serve, ok := h.matchAndHold(req.Messages)
	if !ok {
		writeError(w, "no matching exchange", "no_match", http.StatusBadRequest)
		return
	}
	h.writeAgentEvents(serve.agentEvents)

	stream, _ := rawBody["stream"].(bool)
	if serve.fromLegacy {
		writeError(w, "anthropic endpoint requires generated preset queue", "no_match", http.StatusBadRequest)
		return
	}
	if stream {
		if err := anthropicenc.WriteMessagesStream(w, model, serve.batch); err != nil {
			writeError(w, err.Error(), "internal_error", http.StatusInternalServerError)
		}
		return
	}
	anthropicenc.WriteMessagesNonStream(w, model, serve.batch)
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

type serveResult struct {
	legacy      *mockconfig.Exchange
	batch       queue.Batch
	agentEvents []types.AgentEvent
	fromLegacy  bool
}

func (h *mockHandler) matchAndHold(messages []openai.ChatCompletionMessageParamUnion) (*serveResult, bool) {
	serve, ok := h.findMatch(messages)
	if ok {
		queue.SleepFor(serve.agentEvents)
	}
	return serve, ok
}

func (h *mockHandler) findMatch(messages []openai.ChatCompletionMessageParamUnion) (*serveResult, bool) {
	for i := range h.exchanges {
		ex := h.exchanges[i].Exchange

		if h.counter != h.effectiveIndices[i] {
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
		return &serveResult{
			legacy:      &ex,
			agentEvents: mockgen.ExchangeResponseToAgentEvents(ex.Response),
			fromLegacy:  true,
		}, true
	}

	if h.counter < len(h.exchanges) {
		return nil, false
	}

	return h.findGeneratedMatch(messages)
}

func (h *mockHandler) findGeneratedMatch(messages []openai.ChatCompletionMessageParamUnion) (*serveResult, bool) {
	prompt := extractPrompt(messages)

	h.genMu.Lock()
	if batch, ok := queue.DequeueToBreakpoint(&h.genQueue); ok {
		h.genMu.Unlock()
		batch = sanitizeBatch(batch)
		return &serveResult{
			batch:       batch,
			agentEvents: queue.ConsumedEvents(batch),
		}, true
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
			return &serveResult{
				batch:       batch,
				agentEvents: queue.ConsumedEvents(batch),
			}, true
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
		messages = append(messages, map[string]any{
			"role":    role,
			"content": content,
		})
	}
	return messages
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
