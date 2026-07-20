package view

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	frontend "github.com/xhd2015/agent-pro/frontend-agent-run"
	"github.com/xhd2015/agent-pro/pkgs/assets"
)

// DefaultViewPort is the preferred listen port for grok session view --web.
// If busy, ServeWeb tries the next ports up to ViewPortAttempts total.
const (
	DefaultViewPort  = 61781
	ViewPortAttempts = 100
)

// WebOptions configures the read-only local viewer.
type WebOptions struct {
	// Port is the preferred listen port. Zero or negative uses DefaultViewPort.
	// If that port is taken, the next free port is tried (up to ViewPortAttempts).
	Port int
	// Open opens the session URL in a browser. Default is false (print URL only).
	Open bool
	// Stderr receives status lines (listen URL). Defaults to os.Stderr.
	Stderr io.Writer
}

// ServeWeb starts an agent-run-compatible API + SPA for one in-memory session.
// It bootstraps the viewer, follows updates.jsonl, blocks until interrupted.
func ServeWeb(ctx context.Context, v *Viewer, opts WebOptions) error {
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}

	followCtx, stopFollow := context.WithCancel(ctx)
	defer stopFollow()
	go func() {
		_ = v.Follow(followCtx)
	}()

	mux := http.NewServeMux()
	registerViewAPI(mux, v)
	if err := registerViewStatic(mux, v); err != nil {
		return err
	}

	ln, err := listenLocalPreferred(opts.Port)
	if err != nil {
		return err
	}
	actualPort := ln.Addr().(*net.TCPAddr).Port
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", actualPort)
	sessionURL := baseURL + "/sessions/" + v.Info.ID
	fmt.Fprintf(opts.Stderr, "Serving Grok session view on %s\n", sessionURL)
	fmt.Fprintf(opts.Stderr, "Press Ctrl+C to stop.\n")

	if opts.Open {
		openLocalURL(sessionURL)
	}

	srv := &http.Server{Handler: mux}
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve(ln)
	}()

	sigCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case <-sigCtx.Done():
		stopFollow()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
		<-errCh
		return nil
	case err := <-errCh:
		stopFollow()
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}

func registerViewAPI(mux *http.ServeMux, v *Viewer) {
	mux.HandleFunc("/api/agent-run/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	mux.HandleFunc("/api/agent-run/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		ws := ""
		if v.Info != nil {
			ws = v.Info.CWD
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"home":              "",
			"workspace":         ws,
			"agent_run_home":    "",
			"recent_workspaces": []string{},
			"process_cwd":       ws,
		})
	})

	mux.HandleFunc("/api/agent-run/runners", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"runners": []string{"grok-tty"},
			"default": "grok-tty",
		})
	})

	mux.HandleFunc("/api/agent-run/sessions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			http.Error(w, "read-only: cannot create sessions", http.StatusMethodNotAllowed)
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		sess := v.SessionSummary()
		writeJSON(w, http.StatusOK, map[string]any{
			"sessions": []any{sess},
			"total":    1,
			"limit":    0,
			"offset":   0,
			"has_more": false,
			"counts":   map[string]int{"all": 1, "running": 1, "done": 0},
		})
	})

	mux.HandleFunc("/api/agent-run/sessions/", func(w http.ResponseWriter, r *http.Request) {
		handleViewSessionResource(w, r, v)
	})

	// Workspace / fs not needed for view; return safe stubs.
	mux.HandleFunc("/api/agent-run/workspace", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "read-only viewer", http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/api/agent-run/fs/list", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "read-only viewer", http.StatusMethodNotAllowed)
	})
}

func handleViewSessionResource(w http.ResponseWriter, r *http.Request, v *Viewer) {
	rest := strings.TrimPrefix(r.URL.Path, "/api/agent-run/sessions/")
	rest = strings.Trim(rest, "/")
	parts := strings.Split(rest, "/")
	if len(parts) < 1 || strings.TrimSpace(parts[0]) == "" {
		http.NotFound(w, r)
		return
	}
	sessionID := parts[0]
	if v.Info == nil || sessionID != v.Info.ID {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}

	if len(parts) == 2 && parts[1] == "terminal" {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"available":  false,
			"runner":     "grok-tty",
			"session_id": sessionID,
		})
		return
	}

	if len(parts) == 3 && parts[1] == "events" && parts[2] == "stream" {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handleViewSSE(w, r, v)
		return
	}

	if len(parts) == 2 && parts[1] == "messages" {
		http.Error(w, "read-only: cannot send messages", http.StatusMethodNotAllowed)
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
	writeJSON(w, http.StatusOK, v.DetailPayload())
}

func handleViewSSE(w http.ResponseWriter, r *http.Request, v *Viewer) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	after := int64(0)
	if raw := strings.TrimSpace(r.URL.Query().Get("after")); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
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

	ctx := r.Context()
	ch, cancelSub := v.Subscribe()
	defer cancelSub()

	sendFrom := func(offset int64) (int64, error) {
		data := v.NDJSONAfter(offset)
		if len(data) == 0 {
			return offset, nil
		}
		// data is a slice of the buffer; newOffset = offset + len(data)
		newOffset := offset + int64(len(data))
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if _, err := fmt.Fprintf(w, "data: %s\n\n", line); err != nil {
				return offset, err
			}
			if err := safeFlush(flusher); err != nil {
				return offset, err
			}
		}
		return newOffset, nil
	}

	cur, err := sendFrom(after)
	if err != nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-ch:
			if !ok {
				return
			}
			next, err := sendFrom(cur)
			if err != nil {
				return
			}
			cur = next
		}
	}
}

func safeFlush(flusher http.Flusher) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("sse flush: %v", r)
		}
	}()
	flusher.Flush()
	return nil
}

func registerViewStatic(mux *http.ServeMux, v *Viewer) error {
	dist, err := resolveAgentRunDist()
	if err != nil {
		// Minimal shell when assets incomplete.
		shell := []byte(incompleteHTML())
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write(shell)
		})
		return nil
	}

	indexHTML, err := fs.ReadFile(dist, "index.html")
	if err != nil {
		return err
	}

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
			path = "index.html"
			f, err = dist.Open(path)
			if err != nil {
				http.NotFound(w, r)
				return
			}
		}
		defer f.Close()
		if path != "index.html" {
			w.Header().Set("Content-Type", mimeTypeByExtension(filepath.Ext(path)))
			_, _ = io.Copy(w, f)
			return
		}

		body := indexHTML
		if bootstrap := viewBootstrapJSON(v, r.URL.Path); bootstrap != "" {
			injection := `<script id="agent-run-session-bootstrap" type="application/json">` + bootstrap + `</script>`
			body = bytes.Replace(indexHTML, []byte("</body>"), []byte(injection+"</body>"), 1)
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(body)
	})
	return nil
}

func viewBootstrapJSON(v *Viewer, urlPath string) string {
	parts := strings.Split(strings.Trim(urlPath, "/"), "/")
	if len(parts) != 2 || parts[0] != "sessions" {
		return ""
	}
	if v.Info == nil || strings.TrimSpace(parts[1]) != v.Info.ID {
		return ""
	}
	data, err := json.Marshal(v.DetailPayload())
	if err != nil {
		return ""
	}
	return string(data)
}

func resolveAgentRunDist() (fs.FS, error) {
	version := assets.ClientVersion()
	if frontend.DistComplete() {
		sub, err := fs.Sub(frontend.DistFS, "dist")
		if err == nil {
			return sub, nil
		}
	}
	if assets.CacheComplete(assets.ProductAgentRun, version, assets.KindFrontend) {
		dir, err := assets.AssetCacheDir(assets.ProductAgentRun, version, assets.KindFrontend)
		if err == nil {
			return os.DirFS(dir), nil
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	dir, err := assets.EnsureAsset(ctx, assets.ProductAgentRun, version, assets.KindFrontend, assets.EnsureConfig{})
	if err != nil {
		return nil, err
	}
	return os.DirFS(dir), nil
}

func incompleteHTML() string {
	return `<!DOCTYPE html><html><head><meta charset="utf-8"/><title>Grok session view</title></head>
<body><h1>Frontend assets incomplete</h1>
<p>
  <code>agent-pro grok session view --web</code> needs a fat
  <code>frontend-agent-run</code> embed (message cards). From the repo root:
</p>
<pre>go run ./script/agent-pro/install
# or bundle only, then rebuild agent-pro:
go run ./script/agent-pro/bundle
go install ./cmd/agent-pro</pre>
<p>Alternatively hydrate at runtime: <code>agent-run assets ensure</code>
(with <code>AGENT_PRO_ASSET_BASE_URL</code> set).</p>
</body></html>`
}

// listenLocalPreferred binds 127.0.0.1 starting at preferred (or DefaultViewPort),
// trying the next port on bind failure for up to ViewPortAttempts attempts.
func listenLocalPreferred(preferred int) (net.Listener, error) {
	start := preferred
	if start <= 0 {
		start = DefaultViewPort
	}
	var lastErr error
	for i := 0; i < ViewPortAttempts; i++ {
		port := start + i
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			return ln, nil
		}
		lastErr = err
	}
	end := start + ViewPortAttempts - 1
	return nil, fmt.Errorf("no free port in %d..%d: %w", start, end, lastErr)
}

func mimeTypeByExtension(ext string) string {
	switch ext {
	case ".js":
		return "text/javascript; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".html":
		return "text/html; charset=utf-8"
	case ".svg":
		return "image/svg+xml"
	default:
		return "application/octet-stream"
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func openLocalURL(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
