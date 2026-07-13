// mobile-terminal-touch-scroll: reproduce / verify mobile touch pan scrolling
// of the terminal modal (xterm scrollback) via playwright-debug mobile viewport.
//
// Modes:
//
//	(default) repro  — expect touch pan does NOT reveal older LINE_xxx history.
//	                   Exit 1 + REPRO: lines when symptom present (loop RED).
//	--expect=healthy — expect touch pan scrolls history (DIY touch→scrollLines).
//	                   Exit 0 only when green (VERIFY gate after fix).
//
// Usage (from repo root):
//
//	go run ./script/debug/mobile-terminal-touch-scroll
//	go run ./script/debug/mobile-terminal-touch-scroll --expect=healthy
//	go run ./script/debug/mobile-terminal-touch-scroll -out=/tmp/term-touch
//
// Requires: go, bun, playwright-debug on PATH.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

const (
	sessionID         = "web_term_touch_scroll"
	sessionPath       = "/sessions/" + sessionID
	runnerID          = "codex-tty"
	terminalSessionID = "session-touch-1"
	// Enough lines that mobile viewport (~20–30 rows) cannot show all.
	scrollbackLines = 120
)

type probeReport struct {
	BaseURL        string         `json:"baseURL"`
	SessionPath    string         `json:"sessionPath"`
	SymptomPresent bool           `json:"symptomPresent"`
	OkHealthy      bool           `json:"okHealthy"`
	Issues         []string       `json:"issues"`
	Symptom        map[string]any `json:"symptom"`
	Snapshots      map[string]any `json:"snapshots"`
	Screenshots    []string       `json:"screenshots"`
	Page           map[string]any `json:"page"`
	Gesture        map[string]any `json:"gesture"`
}

func main() {
	expect := flag.String("expect", "repro", "repro | healthy")
	port := flag.Int("port", 0, "web listen port (0 = free ephemeral)")
	outDir := flag.String("out", "", "output dir (default /tmp/mobile-terminal-touch-scroll-<ts>)")
	skipBuild := flag.Bool("skip-build", false, "reuse binaries under -out/bin if present")
	token := flag.String("token", "", "optional API token")
	flag.Parse()

	mode := strings.ToLower(strings.TrimSpace(*expect))
	if mode != "repro" && mode != "healthy" {
		fmt.Fprintf(os.Stderr, "FAIL: --expect must be repro or healthy\n")
		os.Exit(2)
	}

	repoRoot, err := findRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(2)
	}

	if *outDir == "" {
		*outDir = filepath.Join(os.TempDir(), fmt.Sprintf("mobile-terminal-touch-scroll-%d", time.Now().Unix()))
	}
	if err := os.MkdirAll(*outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: mkdir out: %v\n", err)
		os.Exit(2)
	}
	fmt.Fprintf(os.Stderr, "mobile-terminal-touch-scroll: out=%s expect=%s\n", *outDir, mode)

	if _, err := exec.LookPath("playwright-debug"); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: playwright-debug not on PATH: %v\n", err)
		os.Exit(2)
	}
	if _, err := exec.LookPath("bun"); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: bun not on PATH: %v\n", err)
		os.Exit(2)
	}

	binDir := filepath.Join(*outDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(2)
	}
	agentRun := filepath.Join(binDir, "agent-run")

	if !*skipBuild {
		fmt.Fprintln(os.Stderr, "building frontend-agent-run...")
		{
			cmd := exec.Command("bun", "run", "build")
			cmd.Dir = filepath.Join(repoRoot, "frontend-agent-run")
			if out, err := cmd.CombinedOutput(); err != nil {
				fmt.Fprintf(os.Stderr, "FAIL: frontend-agent-run build: %v\n%s\n", err, out)
				os.Exit(2)
			}
		}
		fmt.Fprintln(os.Stderr, "building agent-run...")
		if err := goBuild(repoRoot, "./cmd/agent-run", agentRun); err != nil {
			fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
			os.Exit(2)
		}
	} else if _, err := os.Stat(agentRun); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: missing binary %s (drop -skip-build)\n", agentRun)
		os.Exit(2)
	}

	// Fake ptywrap that streams many LINE_xxx rows for scrollback.
	ptyAddr, ptyCleanup, err := startScrollbackPty()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: start fake pty: %v\n", err)
		os.Exit(2)
	}
	defer ptyCleanup()
	fmt.Fprintf(os.Stderr, "fake ptywrap at %s (lines=%d)\n", ptyAddr, scrollbackLines)

	home := filepath.Join(*outDir, "home", ".agent-run")
	if err := seedTerminalSession(home, runnerID, sessionID, terminalSessionID, ptyAddr); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: seed session: %v\n", err)
		os.Exit(2)
	}

	listenPort := *port
	if listenPort == 0 {
		p, err := freePort()
		if err != nil {
			fmt.Fprintf(os.Stderr, "FAIL: free port: %v\n", err)
			os.Exit(2)
		}
		listenPort = p
	} else if !portFree(listenPort) {
		p, err := freePort()
		if err != nil {
			fmt.Fprintf(os.Stderr, "FAIL: port busy: %v\n", err)
			os.Exit(2)
		}
		fmt.Fprintf(os.Stderr, "port %d busy, using %d\n", *port, p)
		listenPort = p
	}
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", listenPort)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	env := append(os.Environ(), "AGENT_RUN_HOME="+home)
	args := []string{"web", "--port", fmt.Sprint(listenPort), "--no-open", "--agent-runner", runnerID}
	var webStderr bytes.Buffer
	webCmd := exec.CommandContext(ctx, agentRun, args...)
	webCmd.Env = env
	webCmd.Stdout = os.Stdout
	webCmd.Stderr = io.MultiWriter(os.Stderr, &webStderr)
	fmt.Fprintf(os.Stderr, "starting agent-run web on %s ...\n", baseURL)
	if err := webCmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: start web: %v\n", err)
		os.Exit(2)
	}
	stopWeb := func() {
		if webCmd != nil && webCmd.Process != nil {
			_ = webCmd.Process.Signal(syscall.SIGTERM)
			_, _ = webCmd.Process.Wait()
			webCmd = nil
		}
	}
	defer stopWeb()

	if err := waitHealth(ctx, baseURL, *token, 45*time.Second); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: health: %v\nweb stderr:\n%s\n", err, webStderr.String())
		os.Exit(2)
	}
	fmt.Fprintln(os.Stderr, "health: ok")

	// Sanity: terminal status available
	if err := waitTerminalAvailable(ctx, baseURL, sessionID, *token, 15*time.Second); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: terminal status: %v\n", err)
		os.Exit(2)
	}
	fmt.Fprintln(os.Stderr, "terminal: available")

	probeJS := filepath.Join(repoRoot, "script/debug/mobile-terminal-touch-scroll/probe.js")
	probeOut := filepath.Join(*outDir, "probe")
	_ = os.MkdirAll(probeOut, 0755)

	report, raw, err := runProbe(ctx, probeJS, baseURL, probeOut, *token, sessionPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: playwright probe: %v\nraw:\n%s\n", err, raw)
		os.Exit(2)
	}
	_ = os.WriteFile(filepath.Join(*outDir, "probe-report.json"), []byte(mustJSON(report)), 0644)

	fmt.Fprintf(os.Stderr, "probe: symptomPresent=%v okHealthy=%v issues=%v\n",
		report.SymptomPresent, report.OkHealthy, report.Issues)
	if report.Symptom != nil {
		if reason, ok := report.Symptom["reason"].(string); ok && reason != "" {
			fmt.Fprintf(os.Stderr, "probe symptom: %s\n", reason)
		}
	}
	for _, s := range report.Screenshots {
		fmt.Fprintf(os.Stderr, "screenshot: %s\n", s)
	}

	switch mode {
	case "repro":
		if report.SymptomPresent {
			fmt.Println("REPRO: mobile touch pan does not scroll terminal history")
			if reason, ok := report.Symptom["reason"].(string); ok && reason != "" {
				fmt.Printf("REPRO: %s\n", reason)
			}
			for _, issue := range report.Issues {
				fmt.Printf("REPRO: issue: %s\n", issue)
			}
			for _, s := range report.Screenshots {
				fmt.Printf("REPRO: screenshot %s\n", s)
			}
			fmt.Printf("REPRO: report %s\n", filepath.Join(*outDir, "probe-report.json"))
			stopWeb()
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "FAIL: touch-scroll-broken symptom not observed (cannot establish REPRO)\n")
		fmt.Fprintf(os.Stderr, "issues: %v\nokHealthy=%v\n", report.Issues, report.OkHealthy)
		fmt.Fprintf(os.Stderr, "report: %s\n", filepath.Join(*outDir, "probe-report.json"))
		if report.OkHealthy {
			fmt.Fprintln(os.Stderr, "hint: touch scroll already works; use --expect=healthy for VERIFY")
		}
		stopWeb()
		os.Exit(2)

	case "healthy":
		if report.SymptomPresent || !report.OkHealthy {
			fmt.Println("FAIL: mobile touch still does not scroll terminal (not healthy)")
			if reason, ok := report.Symptom["reason"].(string); ok && reason != "" {
				fmt.Printf("FAIL: %s\n", reason)
			}
			for _, issue := range report.Issues {
				fmt.Printf("FAIL: %s\n", issue)
			}
			for _, s := range report.Screenshots {
				fmt.Printf("FAIL: screenshot %s\n", s)
			}
			stopWeb()
			os.Exit(1)
		}
		fmt.Println("VERIFY: mobile touch pan scrolls terminal history")
		for _, s := range report.Screenshots {
			fmt.Printf("VERIFY: screenshot %s\n", s)
		}
		fmt.Printf("VERIFY: report %s\n", filepath.Join(*outDir, "probe-report.json"))
		stopWeb()
		os.Exit(0)
	}
}

func startScrollbackPty() (listenAddr string, cleanup func(), err error) {
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/terminal", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		requestedID := r.URL.Query().Get("session_id")
		if requestedID == "" {
			requestedID = terminalSessionID
		}
		_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"session_id","session_id":"`+requestedID+`"}`))
		// Stream scrollback lines (binary like real PTY).
		var b strings.Builder
		for i := 0; i < scrollbackLines; i++ {
			fmt.Fprintf(&b, "LINE_%03d scrollback filler row for mobile touch pan testing\r\n", i)
		}
		_ = conn.WriteMessage(websocket.BinaryMessage, []byte(b.String()))
		// Keep connection open; echo binary input.
		for {
			mt, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if mt == websocket.BinaryMessage {
				_ = conn.WriteMessage(websocket.BinaryMessage, []byte("echo:"+string(msg)))
			}
		}
	})
	mux.HandleFunc("/api/terminal/sessions", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"id":"` + terminalSessionID + `","status":"running"}]`))
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", nil, err
	}
	srv := &http.Server{Handler: mux}
	go func() { _ = srv.Serve(ln) }()
	addr := ln.Addr().String()
	return addr, func() {
		_ = srv.Close()
		_ = ln.Close()
	}, nil
}

func seedTerminalSession(home, runner, chatID, termID, listenAddr string) error {
	sessDir := filepath.Join(home, "sessions", chatID)
	if err := os.MkdirAll(sessDir, 0755); err != nil {
		return err
	}
	// Flat layout marker so store does not expect nested runner dirs.
	if err := os.WriteFile(filepath.Join(home, "sessions", ".layout"), []byte(`{"version":2}`+"\n"), 0644); err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	meta := map[string]any{
		"runner":              runner,
		"session_id":          chatID,
		"terminal_session_id": termID,
		"status":              "finished",
		"created_at":          now,
		"updated_at":          now,
		"workspace":           filepath.Join(home, "workspace"),
	}
	metaBytes, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(sessDir, "meta.json"), metaBytes, 0644); err != nil {
		return err
	}
	events := `{"type":"message","role":"user","text":"show terminal scrollback","timestamp":1}` + "\n" +
		`{"type":"message","role":"assistant","text":"open the Terminal button to view PTY","timestamp":2}` + "\n"
	if err := os.WriteFile(filepath.Join(sessDir, "events.jsonl"), []byte(events), 0644); err != nil {
		return err
	}

	regDir := filepath.Join(home, runner+"-registry")
	if err := os.MkdirAll(regDir, 0755); err != nil {
		return err
	}
	entry := map[string]any{
		"session_id":  termID,
		"listen_addr": listenAddr,
		"pid":         os.Getpid(),
		"created_at":  time.Now().Format(time.RFC3339Nano),
	}
	entryBytes, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(regDir, termID+".json"), entryBytes, 0644)
}

func waitTerminalAvailable(ctx context.Context, baseURL, sessionID, bearer string, timeout time.Duration) error {
	client := &http.Client{Timeout: 3 * time.Second}
	url := strings.TrimRight(baseURL, "/") + "/api/agent-run/sessions/" + sessionID + "/terminal"
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		if bearer != "" {
			req.Header.Set("Authorization", "Bearer "+bearer)
		}
		resp, err := client.Do(req)
		if err == nil {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK && strings.Contains(string(body), `"available":true`) {
				return nil
			}
		}
		time.Sleep(150 * time.Millisecond)
	}
	return fmt.Errorf("terminal not available: %s", url)
}

func runProbe(ctx context.Context, probeJS, baseURL, outDir, token, sessPath string) (probeReport, string, error) {
	var zero probeReport
	cmd := exec.CommandContext(ctx, "playwright-debug", "run", probeJS, baseURL, outDir, token, sessPath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	if err := cmd.Run(); err != nil {
		return zero, stdout.String() + "\n" + stderr.String(), fmt.Errorf("%w", err)
	}
	line := extractReportJSON(stdout.String())
	if line == "" {
		return zero, stdout.String(), fmt.Errorf("no REPORT_JSON in playwright output")
	}
	var report probeReport
	if err := json.Unmarshal([]byte(line), &report); err != nil {
		return zero, line, fmt.Errorf("parse report: %w", err)
	}
	return report, line, nil
}

func extractReportJSON(stdout string) string {
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "REPORT_JSON=") {
			return strings.TrimPrefix(line, "REPORT_JSON=")
		}
	}
	return ""
}

func waitHealth(ctx context.Context, baseURL, bearer string, timeout time.Duration) error {
	client := &http.Client{Timeout: 2 * time.Second}
	url := strings.TrimRight(baseURL, "/") + "/api/agent-run/health"
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		if bearer != "" {
			req.Header.Set("Authorization", "Bearer "+bearer)
		}
		resp, err := client.Do(req)
		if err == nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(150 * time.Millisecond)
	}
	return fmt.Errorf("health check timed out: %s", url)
}

func goBuild(repoRoot, pkg, out string) error {
	cmd := exec.Command("go", "build", "-o", out, pkg)
	cmd.Dir = repoRoot
	if b, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go build %s: %w\n%s", pkg, err, b)
	}
	return nil
}

func freePort() (int, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	p := ln.Addr().(*net.TCPAddr).Port
	_ = ln.Close()
	return p, nil
}

func portFree(p int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", p))
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from cwd")
		}
		dir = parent
	}
}

func mustJSON(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(b)
}
