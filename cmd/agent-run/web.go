package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentsync"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/less-gen/flags"
)

const webHelp = `
Usage: agent-run web [OPTIONS]

Options:
  --port PORT     listen port (default: 8192, 0 = OS-assigned)
  --token TOKEN   API bearer token (omit = open API; "auto" = generate + require auth)
  --dev           proxy to vite dev server
  --no-open       do not open browser
  --agent-runner RUNNER
                  default agent runner for new web sessions
  --grok-home PATH
                  grok data directory (e.g. ~/.mock/grok); sets GROK_HOME on child processes
  --grok-tty-runner-binary SPEC
                  grok-tty binary name/path for web runs (e.g. llm-mock-run-grok)
  -h, --help      show help
`

func runWeb(args []string, defaultRunner string) error {
	var port int = 8192
	var token string
	var dev bool
	var noOpen bool
	var agentRunner string
	var grokHome string
	var grokTTYRunnerBinary string
	_, err := flags.Int("--port", &port).
		String("--token", &token).
		Bool("--dev", &dev).
		Bool("--no-open", &noOpen).
		String("--agent-runner", &agentRunner).
		String("--grok-home", &grokHome).
		String("--grok-tty-runner-binary", &grokTTYRunnerBinary).
		Help("-h,--help", webHelp).
		Parse(args)
	if err != nil {
		return err
	}
	_ = dev
	_ = noOpen

	expandedGrokHome, err := expandHomePath(grokHome)
	if err != nil {
		return err
	}
	if expandedGrokHome != "" {
		if err := os.MkdirAll(expandedGrokHome, 0755); err != nil {
			return fmt.Errorf("create grok home %s: %w", expandedGrokHome, err)
		}
	}

	runCfg := webRunConfig{
		GrokHome:            expandedGrokHome,
		GrokTTYRunnerBinary: strings.TrimSpace(grokTTYRunnerBinary),
		DefaultRunner:       strings.TrimSpace(agentRunner),
	}
	if runCfg.DefaultRunner == "" {
		runCfg.DefaultRunner = strings.TrimSpace(defaultRunner)
	}

	store, err := openStore()
	if err != nil {
		return err
	}
	maybeInstallCodexTTYTestFixture(store.Home())
	mode, _ := webTokenArgMode(args)
	var requireAuth bool
	switch mode {
	case webTokenOmit:
		requireAuth = false
		fmt.Fprintf(os.Stderr, "no API token configured\n")
		fmt.Fprintf(os.Stderr, "--token <secret> or --token auto to require Bearer authentication\n")
	case webTokenAuto:
		requireAuth = true
		token, err = generateWebToken()
		if err != nil {
			return err
		}
		if err := writeAuthToken(store.Home(), token); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "agent-run web token: %s\n", token)
	case webTokenExplicit:
		requireAuth = true
		token = strings.TrimSpace(token)
		if token == "" {
			return fmt.Errorf("--token requires a value (or use --token auto)")
		}
		if err := writeAuthToken(store.Home(), token); err != nil {
			return err
		}
	}

	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return err
	}
	actualPort := ln.Addr().(*net.TCPAddr).Port
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", actualPort)
	fmt.Fprintf(os.Stderr, "agent-run web listening at %s\n", baseURL)

	mux := http.NewServeMux()
	registerAPI(mux, store, token, requireAuth, runCfg)
	if err := registerStatic(mux, store); err != nil {
		return err
	}

	reconcileGrokHome := runCfg.GrokHome
	if reconcileGrokHome == "" {
		reconcileGrokHome = agenttty.GrokHome()
	}
	agentsync.StartReconciler(context.Background(), 5*time.Minute, agentsync.ReconcileOptions{
		Home:     store.Home(),
		GrokHome: reconcileGrokHome,
		Runner:   "grok-tty",
	})

	srv := &http.Server{Handler: mux}
	return srv.Serve(ln)
}

type webTokenMode int

const (
	webTokenOmit webTokenMode = iota
	webTokenAuto
	webTokenExplicit
)

func webTokenArgMode(args []string) (webTokenMode, string) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--token" {
			if i+1 >= len(args) {
				return webTokenExplicit, ""
			}
			v := args[i+1]
			if v == "auto" {
				return webTokenAuto, ""
			}
			return webTokenExplicit, v
		}
		if strings.HasPrefix(arg, "--token=") {
			v := strings.TrimPrefix(arg, "--token=")
			if v == "auto" {
				return webTokenAuto, ""
			}
			return webTokenExplicit, v
		}
	}
	return webTokenOmit, ""
}

func generateWebToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func writeAuthToken(home, token string) error {
	if err := os.MkdirAll(home, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(home, "auth.token"), []byte(token), 0600)
}

func isTerminalWebSocketPath(path string) bool {
	path = strings.TrimSuffix(strings.Trim(path, "/"), "/ws")
	return strings.HasSuffix(path, "/terminal")
}

func registerAPI(mux *http.ServeMux, store agentstorage.Store, token string, requireAuth bool, runCfg webRunConfig) {
	wrap := func(next http.HandlerFunc) http.HandlerFunc {
		if !requireAuth {
			return next
		}
		return func(w http.ResponseWriter, r *http.Request) {
			got := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			got = strings.TrimSpace(got)
			if got == "" {
				got = strings.TrimSpace(r.URL.Query().Get("token"))
			}
			if got != "" && got != token {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if got == "" && !isTerminalWebSocketPath(r.URL.Path) {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next(w, r)
		}
	}
	auth := wrap
	mux.HandleFunc("/api/agent-run/health", auth(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	mux.HandleFunc("/api/agent-run/status", auth(handleStatus(store)))
	mux.HandleFunc("/api/agent-run/workspace", auth(handleWorkspace(store)))
	mux.HandleFunc("/api/agent-run/fs/list", auth(handleFSList()))
	mux.HandleFunc("/api/agent-run/runners", auth(handleRunners(store, runCfg)))
	mux.HandleFunc("/api/agent-run/sessions", auth(handleSessionsCollection(store, runCfg)))
	mux.HandleFunc("/api/agent-run/sessions/", auth(handleSessionResource(store, runCfg)))
}

func handleStatus(store agentstorage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		payload, err := statusPayload(store)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, payload)
	}
}

func handleWorkspace(store agentstorage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Path string `json:"path"`
		}
		if err := readJSONBody(r, &body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		path := strings.TrimSpace(body.Path)
		if path == "" {
			http.Error(w, "path is required", http.StatusBadRequest)
			return
		}
		if !filepath.IsAbs(path) {
			http.Error(w, "path must be absolute", http.StatusBadRequest)
			return
		}
		path = canonicalWorkspacePath(path)
		st, err := os.Stat(path)
		if err != nil {
			http.Error(w, "path does not exist", http.StatusBadRequest)
			return
		}
		if !st.IsDir() {
			http.Error(w, "path must be a directory", http.StatusBadRequest)
			return
		}
		cfg, err := store.Config()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		cfg.SelectedWorkspace = path
		cfg.RecentWorkspaces = pushRecentWorkspace(cfg.RecentWorkspaces, path, maxRecentWorkspaces)
		if err := store.SaveConfig(cfg); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		payload, err := statusPayload(store)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, payload)
	}
}

func handleFSList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		path := strings.TrimSpace(r.URL.Query().Get("path"))
		if path == "" {
			http.Error(w, "path is required", http.StatusBadRequest)
			return
		}
		if !filepath.IsAbs(path) {
			http.Error(w, "path must be absolute", http.StatusBadRequest)
			return
		}
		path = canonicalWorkspacePath(path)
		st, err := os.Stat(path)
		if err != nil {
			http.Error(w, "path does not exist", http.StatusBadRequest)
			return
		}
		if !st.IsDir() {
			http.Error(w, "path must be a directory", http.StatusBadRequest)
			return
		}
		entries, err := os.ReadDir(path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		type fsEntry struct {
			Name string `json:"name"`
			Path string `json:"path"`
			Type string `json:"type"`
		}
		out := make([]fsEntry, 0, len(entries))
		for _, e := range entries {
			name := e.Name()
			// Include dot names (.git, .env, etc.) so the browser can navigate hidden dirs/files.
			full := canonicalWorkspacePath(filepath.Join(path, name))
			typ := "file"
			if e.IsDir() {
				typ = "dir"
			} else if e.Type()&os.ModeSymlink != 0 {
				// Resolve symlink for type classification.
				if info, err := os.Stat(full); err == nil && info.IsDir() {
					typ = "dir"
				}
			}
			out = append(out, fsEntry{Name: name, Path: full, Type: typ})
		}
		// Dirs first, then files; case-insensitive name within each group.
		sort.SliceStable(out, func(i, j int) bool {
			iDir := out[i].Type == "dir"
			jDir := out[j].Type == "dir"
			if iDir != jDir {
				return iDir // dirs before files
			}
			return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
		})
		parent := filepath.Dir(path)
		if parent == path {
			// Root has no parent; omit or use empty.
			parent = ""
		} else {
			parent = canonicalWorkspacePath(parent)
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"path":    path,
			"parent":  parent,
			"entries": out,
		})
	}
}


