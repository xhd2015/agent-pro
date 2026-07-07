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
	sess, err := store.GetSession(runner, sessionID)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(sess.Meta.Workspace) != "" {
		return sess.Meta.Workspace, nil
	}
	return webProcessWorkspace()
}

func appendUserPromptEvent(store agentstorage.Store, runner, sessionID, text string) error {
	return store.AppendEvent(runner, sessionID, types.AgentEvent{
		Type:      types.ActionMessage,
		Role:      "user",
		Text:      text,
		Timestamp: time.Now().UnixMilli(),
	})
}

func waitForLiveTerminal(store agentstorage.Store, runner, sessionID string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if ttySess, err := agenttty.ResolveTerminalStatus(store, runner, sessionID); err == nil && ttySess.TCPReachable {
			if strings.TrimSpace(ttySess.Meta.TerminalSessionID) == "" && ttySess.TerminalSessionID != "" {
				_ = store.UpdateSessionTerminalSessionID(runner, sessionID, ttySess.TerminalSessionID)
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
		startGrokFollowUpEventTail(store, runCfg, runner, sessionID)
	}
	return true
}

func startGrokFollowUpEventTail(store agentstorage.Store, runCfg webRunConfig, runner, sessionID string) {
	meta, err := store.GetSession(runner, sessionID)
	if err != nil {
		return
	}
	grokSessionID := strings.TrimSpace(meta.Meta.RunnerSessionID)
	if grokSessionID == "" {
		grokSessionID = strings.TrimSpace(os.Getenv("AGENT_RUN_GROK_TTY_GROK_SESSION_ID"))
	}
	if grokSessionID == "" {
		return
	}
	workspace, err := sessionWorkspace(store, runner, sessionID)
	if err != nil {
		return
	}
	grokHome := agenttty.GrokHomeForRunner(runCfg.GrokHome)
	startOffset := agenttty.UpdatesTailOffset(grokHome, workspace, grokSessionID)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		emit := func(ev types.AgentEvent) error {
			if ev.Type == types.ActionMessage && strings.TrimSpace(ev.Role) == "" {
				ev.Role = "assistant"
			}
			if ev.Type == types.ActionMessage && ev.Timestamp == 0 {
				ev.Timestamp = time.Now().UnixMilli()
			}
			return store.AppendEvent(runner, sessionID, ev)
		}
		_ = agenttty.TailGrokSessionFromOffset(ctx, grokHome, workspace, grokSessionID, startOffset, emit)
		_ = emit(types.AgentEvent{Type: types.ActionDone})
		_ = store.UpdateSessionStatus(runner, sessionID, "finished")
	}()
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
			Prompt:            prompt,
			Runner:            runner,
			SessionID:         sessionID,
			Workspace:         workspace,
			Store:             store,
			Stdout:            io.Discard,
			Stderr:            io.Discard,
			StreamPhases:      false,
			KeepTerminalAlive: agenttty.IsTTYRunner(runner),
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
			if _, err := store.GetSession(runner, sessionID); err != nil {
				if err := store.CreateSession(runner, sessionID, agentstorage.SessionMeta{
					Runner:    runner,
					SessionID: sessionID,
					Status:    "running",
					Workspace: workspace,
				}); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			} else {
				_ = store.UpdateSessionStatus(runner, sessionID, "running")
			}
			if err := appendUserPromptEvent(store, runner, sessionID, prompt); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			startAgentRun(store, runCfg, runner, sessionID, prompt)
			if agenttty.IsTTYRunner(runner) {
				waitForLiveTerminal(store, runner, sessionID, 15*time.Second)
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
		if len(parts) < 2 {
			http.NotFound(w, r)
			return
		}
		runner, sessionID := parts[0], parts[1]

		if len(parts) == 3 && parts[2] == "terminal" {
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handleTerminalStatus(store, runner, sessionID)(w, r)
			return
		}

		if len(parts) == 4 && parts[2] == "terminal" && parts[3] == "ws" {
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handleTerminalWebSocket(store, runner, sessionID)(w, r)
			return
		}

		if len(parts) == 4 && parts[2] == "events" && parts[3] == "stream" {
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handleSessionEventsStream(store, runner, sessionID)(w, r)
			return
		}

		if len(parts) == 3 && parts[2] == "messages" {
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
			if _, err := store.GetSession(runner, sessionID); err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			_ = store.UpdateSessionStatus(runner, sessionID, "running")
			if err := appendUserPromptEvent(store, runner, sessionID, text); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if sent := enqueueLiveTerminalMessage(store, runCfg, runner, sessionID, text); sent {
				writeJSON(w, http.StatusAccepted, map[string]any{"ok": true})
				return
			}
			startAgentRun(store, runCfg, runner, sessionID, text)
			writeJSON(w, http.StatusAccepted, map[string]any{"ok": true})
			return
		}

		if len(parts) != 2 {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		meta, err := store.GetSession(runner, sessionID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		events, eventsOffset, err := store.ReadEvents(runner, sessionID, 0)
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

func handleSessionEventsStream(store agentstorage.Store, runner, sessionID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := store.GetSession(runner, sessionID); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
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

		_ = agentevents.WatchEvents(r.Context(), store, runner, sessionID, after, func(line string) error {
			if _, err := fmt.Fprintf(w, "data: %s\n\n", line); err != nil {
				return err
			}
			flusher.Flush()
			return nil
		})
	}
}
