package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	frontend "github.com/xhd2015/agent-pro/frontend-agent-run"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
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
  -h, --help      show help
`

func runWeb(args []string, defaultRunner string) error {
	var port int = 8192
	var token string
	var dev bool
	var noOpen bool
	var agentRunner string
	_, err := flags.Int("--port", &port).
		String("--token", &token).
		Bool("--dev", &dev).
		Bool("--no-open", &noOpen).
		String("--agent-runner", &agentRunner).
		Help("-h,--help", webHelp).
		Parse(args)
	if err != nil {
		return err
	}
	_ = defaultRunner
	_ = dev
	_ = noOpen
	_ = agentRunner

	store, err := openStore()
	if err != nil {
		return err
	}
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
	registerAPI(mux, store, token, requireAuth)
	if err := registerStatic(mux); err != nil {
		return err
	}

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

func registerAPI(mux *http.ServeMux, store agentstorage.Store, token string, requireAuth bool) {
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
			if got != token {
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
	mux.HandleFunc("/api/agent-run/status", auth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		workspace, err := webProcessWorkspace()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"home":      store.Home(),
			"workspace": workspace,
		})
	}))
	mux.HandleFunc("/api/agent-run/runners", auth(handleRunners(store)))
	mux.HandleFunc("/api/agent-run/sessions", auth(handleSessionsCollection(store)))
	mux.HandleFunc("/api/agent-run/sessions/", auth(handleSessionResource(store)))
}

func registerStatic(mux *http.ServeMux) error {
	dist, err := fs.Sub(frontend.DistFS, "dist")
	if err != nil {
		return err
	}
	fileServer := http.FileServer(http.FS(dist))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		f, err := dist.Open(path)
		if err != nil {
			// SPA fallback
			path = "index.html"
			f, err = dist.Open(path)
			if err != nil {
				http.NotFound(w, r)
				return
			}
		}
		defer f.Close()
		if path == "index.html" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		} else {
			w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(path)))
		}
		_, _ = io.Copy(w, f)
	})
	_ = fileServer
	return nil
}
