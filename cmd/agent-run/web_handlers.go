package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/xhd2015/agent-pro/agent/cli/registry"
	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/pkgs/agentevents"
	"github.com/xhd2015/agent-pro/pkgs/agentsend"
	"github.com/xhd2015/agent-pro/pkgs/agentsync"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/agent-pro/pkgs/agentui"
)

var webRunWG sync.WaitGroup

// canonicalWorkspacePath matches paths like t.TempDir() on macOS (/var vs /private/var).
func canonicalWorkspacePath(path string) string {
	path = filepath.Clean(path)
	if !filepath.IsAbs(path) {
		if abs, err := filepath.Abs(path); err == nil {
			path = filepath.Clean(abs)
		}
	}
	if strings.HasPrefix(path, "/private/") {
		alt := strings.TrimPrefix(path, "/private")
		if alt != "" && alt[0] == '/' {
			if a, errA := os.Lstat(path); errA == nil {
				if b, errB := os.Lstat(alt); errB == nil && os.SameFile(a, b) {
					return filepath.Clean(alt)
				}
			}
		}
	}
	return path
}

func knownRunners() []string {
	return []string{
		string(registry.AgentRunnerOpencode),
		string(registry.AgentRunnerCodex),
		string(registry.AgentRunnerCodexTTY),
		string(registry.AgentRunnerCursor),
		string(registry.AgentRunnerPi),
		string(registry.AgentRunnerCrush),
		string(registry.AgentRunnerGrok),
		string(registry.AgentRunnerGrokTTY),
		string(registry.AgentRunnerFakeCodex),
	}
}

func defaultRunner(store agentstorage.Store, runCfg webRunConfig) string {
	if strings.TrimSpace(runCfg.DefaultRunner) != "" {
		return runCfg.DefaultRunner
	}
	cfg, err := store.Config()
	if err == nil && strings.TrimSpace(cfg.DefaultAgentRunner) != "" {
		return cfg.DefaultAgentRunner
	}
	return string(registry.AgentRunnerOpencode)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func readJSONBody(r *http.Request, dst any) error {
	defer r.Body.Close()
	data, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return fmt.Errorf("empty request body")
	}
	return json.Unmarshal(data, dst)
}

func newWebSessionID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "web_" + hex.EncodeToString(b)
}

func webProcessWorkspace() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return canonicalWorkspacePath(wd), nil
}

func sessionWorkspace(store agentstorage.Store, runner, sessionID string) (string, error) {
	sess, err := store.GetSession(sessionID)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(sess.Meta.Workspace) != "" {
		return sess.Meta.Workspace, nil
	}
	return webProcessWorkspace()
}

func appendUserPromptEvent(store agentstorage.Store, runner, sessionID, text string) error {
	return store.AppendEvent(sessionID, types.AgentEvent{
		Type:      types.ActionMessage,
		Role:      "user",
		Text:      text,
		Timestamp: time.Now().UnixMilli(),
	})
}

func grokTTYUserMessageFromTail(runner string) bool {
	return runner == string(registry.AgentRunnerGrokTTY)
}

func maybeAppendUserPromptEvent(store agentstorage.Store, runner, sessionID, text string) error {
	if grokTTYUserMessageFromTail(runner) {
		return nil
	}
	return appendUserPromptEvent(store, runner, sessionID, text)
}

func waitForLiveTerminal(store agentstorage.Store, runner, sessionID string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if ttySess, err := agenttty.ResolveTerminalStatus(store, runner, sessionID); err == nil && ttySess.TCPReachable {
			if strings.TrimSpace(ttySess.Meta.TerminalSessionID) == "" && ttySess.TerminalSessionID != "" {
				_ = store.UpdateSessionTerminalSessionID(sessionID, ttySess.TerminalSessionID)
			}
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func enqueueLiveTerminalMessage(store agentstorage.Store, runCfg webRunConfig, runner, sessionID, text string) bool {
	ttySess, err := agenttty.ResolveByAgentSession(store, runner, sessionID)
	if err != nil || !ttySess.TCPReachable {
		return false
	}
	provider, ok := agenttty.Get(runner)
	if !ok {
		return false
	}
	sess := agentsend.Session{
		Home:              store.Home(),
		Runner:            runner,
		TerminalSessionID: ttySess.TerminalSessionID,
		ListenAddr:        ttySess.Registry.ListenAddr,
	}
	if _, err := agentsend.Enqueue(store.Home(), sess, text); err != nil {
		return false
	}
	go agentsend.StartDrainer(store.Home(), sess, provider)
	if runner == "grok-tty" {
		ensureWebGrokSync(runner, sessionID, store, runCfg)
	}
	return true
}

func waitForGrokUpdatesPath(grokHome, workspace, grokSessionID string, timeout time.Duration) string {
	grokSessionID = strings.TrimSpace(grokSessionID)
	if grokSessionID == "" {
		return ""
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if path, ok := agenttty.FindUpdatesBySessionID(grokHome, workspace, grokSessionID); ok {
			if abs, err := filepath.Abs(path); err == nil {
				return abs
			}
			return path
		}
		time.Sleep(50 * time.Millisecond)
	}
	return ""
}

func ensureWebGrokSync(runner, sessionID string, store agentstorage.Store, runCfg webRunConfig, promptHint ...string) {
	meta, err := store.GetSession(sessionID)
	if err != nil {
		return
	}
	grokSessionID := strings.TrimSpace(meta.Meta.RunnerSessionID)
	if grokSessionID == "" {
		grokSessionID = strings.TrimSpace(os.Getenv("AGENT_RUN_GROK_TTY_GROK_SESSION_ID"))
	}
	workspace, err := sessionWorkspace(store, runner, sessionID)
	if err != nil {
		return
	}
	grokHome := agenttty.GrokHomeForRunner(runCfg.GrokHome)
	updatesPath := ""
	if grokSessionID != "" {
		if path, ok := agenttty.FindUpdatesBySessionID(grokHome, workspace, grokSessionID); ok {
			updatesPath = path
			if abs, absErr := filepath.Abs(updatesPath); absErr == nil {
				updatesPath = abs
			}
		} else if len(promptHint) > 0 {
			updatesPath = waitForGrokUpdatesPath(grokHome, workspace, grokSessionID, 12*time.Second)
		}
	}
	initialPrompt := strings.TrimSpace(meta.Meta.InitialPrompt)
	if initialPrompt == "" && len(promptHint) > 0 {
		initialPrompt = strings.TrimSpace(promptHint[0])
	}
	sink := agentsync.NewStoreGrokSyncSink(store, runner, sessionID, grokSessionID, updatesPath)
	createdAt := time.Now().Add(-2 * time.Second)
	if t, parseErr := time.Parse(time.RFC3339Nano, strings.TrimSpace(meta.Meta.CreatedAt)); parseErr == nil {
		createdAt = t
	} else if t, parseErr := time.Parse(time.RFC3339, strings.TrimSpace(meta.Meta.CreatedAt)); parseErr == nil {
		createdAt = t
	}
	_ = agentsync.EnsureGrokSync(context.Background(), agentsync.GrokSyncOptions{
		Runner:           runner,
		SessionID:        sessionID,
		GrokSessionID:    grokSessionID,
		UpdatesPath:      updatesPath,
		Workspace:        workspace,
		GrokHome:         grokHome,
		InitialPrompt:    initialPrompt,
		SessionCreatedAt: createdAt,
		Sink:             sink,
	})
}

func startAgentRun(store agentstorage.Store, runCfg webRunConfig, runner, sessionID, prompt string) {
	workspace, _ := sessionWorkspace(store, runner, sessionID)
	webRunWG.Add(1)
	go func() {
		defer webRunWG.Done()
		if _, err := agentsend.SendToAgentSession(store, runner, sessionID, prompt, agentsend.WaitOptions{
			Mode:         agentsend.WaitNoWait,
			StartDrainer: true,
		}); err == nil {
			return
		}
		runOpts := agentui.RunOptions{
			Prompt:             prompt,
			Runner:             runner,
			SessionID:          sessionID,
			Workspace:          workspace,
			Store:              store,
			Stdout:             io.Discard,
			Stderr:             io.Discard,
			StreamPhases:       false,
			KeepTerminalAlive:  agenttty.IsTTYRunner(runner),
			WebManagedGrokSync: runner == "grok-tty",
		}
		applyWebGrokRunOptions(runner, runCfg, &runOpts)
		_ = agentui.Run(context.Background(), runOpts)
	}()
}

type createSessionRequest struct {
	Runner    string `json:"runner"`
	Prompt    string `json:"prompt"`
	SessionID string `json:"session_id,omitempty"`
}

type sendMessageRequest struct {
	Text    string `json:"text"`
	Message string `json:"message"`
}

func handleRunners(store agentstorage.Store, runCfg webRunConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"runners": knownRunners(),
			"default": defaultRunner(store, runCfg),
		})
	}
}

func handleSessionsCollection(store agentstorage.Store, runCfg webRunConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			list, err := listAllSessions(store)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if list == nil {
				list = []agentstorage.SessionMeta{}
			}
			writeJSON(w, http.StatusOK, map[string]any{"sessions": list})
		case http.MethodPost:
			var req createSessionRequest
			if err := readJSONBody(r, &req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			runner := strings.TrimSpace(req.Runner)
			if runner == "" {
				runner = defaultRunner(store, runCfg)
			}
			if err := validateRunner(runner); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			prompt := strings.TrimSpace(req.Prompt)
			if prompt == "" {
				http.Error(w, "prompt is required", http.StatusBadRequest)
				return
			}
			sessionID := strings.TrimSpace(req.SessionID)
			if sessionID == "" {
				sessionID = newWebSessionID()
			}
			workspace, err := webProcessWorkspace()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if _, err := store.GetSession(sessionID); err != nil {
				if err := store.CreateSession(sessionID, agentstorage.SessionMeta{
					Runner:        runner,
					SessionID:     sessionID,
					Status:        "running",
					Workspace:     workspace,
					InitialPrompt: prompt,
				}); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			} else {
				_ = store.UpdateSessionStatus(sessionID, "running")
			}
			if err := maybeAppendUserPromptEvent(store, runner, sessionID, prompt); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			startAgentRun(store, runCfg, runner, sessionID, prompt)
			if agenttty.IsTTYRunner(runner) {
				waitForLiveTerminal(store, runner, sessionID, 15*time.Second)
			}
			if runner == "grok-tty" {
				ensureWebGrokSync(runner, sessionID, store, runCfg)
			}
			writeJSON(w, http.StatusAccepted, map[string]any{
				"session": agentstorage.SessionMeta{
					Runner:    runner,
					SessionID: sessionID,
					Status:    "running",
					Workspace: workspace,
				},
			})
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func handleSessionResource(store agentstorage.Store, runCfg webRunConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rest := strings.TrimPrefix(r.URL.Path, "/api/agent-run/sessions/")
		rest = strings.Trim(rest, "/")
		if rest == "" {
			http.NotFound(w, r)
			return
		}
		parts := strings.Split(rest, "/")
		if len(parts) < 1 || strings.TrimSpace(parts[0]) == "" {
			http.NotFound(w, r)
			return
		}
		sessionID := parts[0]

		// Load runner from meta when session exists; create/list paths still pass runner in body.
		runnerFromMeta := func() (string, *agentstorage.Session, error) {
			meta, err := store.GetSession(sessionID)
			if err != nil {
				return "", nil, err
			}
			return meta.Meta.Runner, meta, nil
		}

		if len(parts) == 2 && parts[1] == "terminal" {
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			runner, _, err := runnerFromMeta()
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			handleTerminalStatus(store, runner, sessionID)(w, r)
			return
		}

		if len(parts) == 3 && parts[1] == "terminal" && parts[2] == "ws" {
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			runner, _, err := runnerFromMeta()
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			handleTerminalWebSocket(store, runner, sessionID)(w, r)
			return
		}

		if len(parts) == 3 && parts[1] == "events" && parts[2] == "stream" {
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			runner, _, err := runnerFromMeta()
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			handleSessionEventsStream(store, runCfg, runner, sessionID)(w, r)
			return
		}

		if len(parts) == 2 && parts[1] == "messages" {
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			var req sendMessageRequest
			if err := readJSONBody(r, &req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			text := strings.TrimSpace(req.Text)
			if text == "" {
				text = strings.TrimSpace(req.Message)
			}
			if text == "" {
				http.Error(w, "text is required", http.StatusBadRequest)
				return
			}
			runner, _, err := runnerFromMeta()
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			_ = store.UpdateSessionStatus(sessionID, "running")
			if err := maybeAppendUserPromptEvent(store, runner, sessionID, text); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if sent := enqueueLiveTerminalMessage(store, runCfg, runner, sessionID, text); sent {
				if runner == "grok-tty" {
					_ = agentsync.StopGrokSync(runner, sessionID)
					ensureWebGrokSync(runner, sessionID, store, runCfg, text)
				}
				writeJSON(w, http.StatusAccepted, map[string]any{"ok": true})
				return
			}
			if runner == "grok-tty" {
				_ = agentsync.StopGrokSync(runner, sessionID)
			}
			startAgentRun(store, runCfg, runner, sessionID, text)
			if agenttty.IsTTYRunner(runner) {
				waitForLiveTerminal(store, runner, sessionID, 15*time.Second)
			}
			if runner == "grok-tty" {
				ensureWebGrokSync(runner, sessionID, store, runCfg, text)
			}
			writeJSON(w, http.StatusAccepted, map[string]any{"ok": true})
			return
		}

		if len(parts) != 1 {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		runner, meta, err := runnerFromMeta()
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if runner == "grok-tty" {
			ensureWebGrokSync(runner, sessionID, store, runCfg)
		}
		events, eventsOffset, err := store.ReadEvents(sessionID, 0)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if events == nil {
			events = []types.AgentEvent{}
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"session":       meta.Meta,
			"events":        events,
			"events_offset": eventsOffset,
		})
	}
}

func safeSSEFlush(flusher http.Flusher) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("sse flush: %v", r)
		}
	}()
	flusher.Flush()
	return nil
}

func handleSessionEventsStream(store agentstorage.Store, runCfg webRunConfig, runner, sessionID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := store.GetSession(sessionID); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if runner == "grok-tty" {
			ensureWebGrokSync(runner, sessionID, store, runCfg)
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming not supported", http.StatusInternalServerError)
			return
		}

		after := int64(0)
		if v := strings.TrimSpace(r.URL.Query().Get("after")); v != "" {
			parsed, err := strconv.ParseInt(v, 10, 64)
			if err != nil || parsed < 0 {
				http.Error(w, "invalid after offset", http.StatusBadRequest)
				return
			}
			after = parsed
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")

		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		eventsPath := filepath.Join(store.Home(), "sessions", sessionID, "events.jsonl")
		tailFromEOF := false
		if info, err := os.Stat(eventsPath); err == nil && after >= info.Size() {
			tailFromEOF = true
		}

		var (
			lastLineMu sync.Mutex
			lastLineAt time.Time
		)
		recordTailLine := func() {
			lastLineMu.Lock()
			lastLineAt = time.Now()
			lastLineMu.Unlock()
		}
		idleSinceLastTailLine := func() time.Duration {
			lastLineMu.Lock()
			defer lastLineMu.Unlock()
			if lastLineAt.IsZero() {
				return 0
			}
			return time.Since(lastLineAt)
		}

		if tailFromEOF {
			go pollWebSSEFinishedEOFExit(ctx, cancel, store, runner, sessionID, idleSinceLastTailLine)
		}

		_ = agentevents.WatchEvents(ctx, store, sessionID, after, func(line string) error {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			recordTailLine()
			if _, err := fmt.Fprintf(w, "data: %s\n\n", line); err != nil {
				return err
			}
			return safeSSEFlush(flusher)
		})
	}
}

// pollWebSSEFinishedEOFExit closes SSE when tailing from EOF on a finished session
// with no new lines, so clients do not block until their own timeout.
func pollWebSSEFinishedEOFExit(ctx context.Context, cancel context.CancelFunc, store agentstorage.Store, runner, sessionID string, idleSinceLastLine func() time.Duration) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			meta, err := store.GetSession(sessionID)
			if err != nil || meta.Meta.Status == "running" {
				continue
			}
			if idle := idleSinceLastLine(); idle == 0 || idle >= sessionsPrintIdleAfterLine {
				cancel()
				return
			}
		}
	}
}
