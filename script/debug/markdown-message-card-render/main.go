// markdown-message-card-render: reproduce / verify markdown rendering on
// thinking progress cards and assistant response message cards.
//
// Modes:
//
//	(default) repro  — expect plain-text markdown markers (not rendered).
//	                   Exit 1 + REPRO: lines when symptom present (loop RED).
//	--expect=healthy — expect rendered markdown that "looks good"
//	                   (strong/pre/code present; no raw ** / ``` in assistant).
//	                   Exit 0 only when green (VERIFY gate after fix).
//
// Optional:
//
//	-url=http://127.0.0.1:8821 -session=/sessions/web_xxx
//	  skip local seed/web start and probe an already-running server.
//
// Usage (from repo root):
//
//	go run ./script/debug/markdown-message-card-render
//	go run ./script/debug/markdown-message-card-render --expect=healthy
//	go run ./script/debug/markdown-message-card-render -url=http://127.0.0.1:8821 -session=/sessions/web_da6177509c41c080
//
// Requires: go, bun (when not using -url), playwright-debug on PATH.
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
)

const (
	sessionID   = "layout-md-render"
	sessionPath = "/sessions/" + sessionID
	// Seed mirrors real grok-tty content shape (bold + fence) plus think markers.
	seedUser = "run ls and pwd"
	seedThink = "Plan: run **pwd** then `ls` and format results for the user."
	seedAssistant = "**pwd:** `/tmp/demo-workspace`\n\n**ls:**\n```\nfile_a.md\nfile_b.go\nsubdir\n```\n\nDone."
)

type probeReport struct {
	BaseURL        string         `json:"baseURL"`
	SessionPath    string         `json:"sessionPath"`
	SymptomPresent bool           `json:"symptomPresent"`
	OkHealthy      bool           `json:"okHealthy"`
	Issues         []string       `json:"issues"`
	Symptom        map[string]any `json:"symptom"`
	Snapshot       map[string]any `json:"snapshot"`
	Screenshots    []string       `json:"screenshots"`
	Page           map[string]any `json:"page"`
}

func main() {
	expect := flag.String("expect", "repro", "repro | healthy")
	port := flag.Int("port", 0, "web listen port (0 = free ephemeral)")
	outDir := flag.String("out", "", "output dir (default /tmp/markdown-message-card-render-<ts>)")
	skipBuild := flag.Bool("skip-build", false, "reuse binaries under -out/bin if present")
	liveURL := flag.String("url", "", "optional existing base URL (skip seed + start web)")
	liveSession := flag.String("session", sessionPath, "session path when using -url")
	token := flag.String("token", "", "optional API token for probe localStorage")
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
		*outDir = filepath.Join(os.TempDir(), fmt.Sprintf("markdown-message-card-render-%d", time.Now().Unix()))
	}
	if err := os.MkdirAll(*outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: mkdir out: %v\n", err)
		os.Exit(2)
	}
	fmt.Fprintf(os.Stderr, "markdown-message-card-render: out=%s expect=%s\n", *outDir, mode)

	if _, err := exec.LookPath("playwright-debug"); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: playwright-debug not on PATH: %v\n", err)
		os.Exit(2)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var baseURL string
	var sessPath string
	var webCmd *exec.Cmd

	if strings.TrimSpace(*liveURL) != "" {
		baseURL = strings.TrimRight(strings.TrimSpace(*liveURL), "/")
		sessPath = *liveSession
		if sessPath == "" {
			sessPath = sessionPath
		}
		fmt.Fprintf(os.Stderr, "using live url %s session %s\n", baseURL, sessPath)
	} else {
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

		home := filepath.Join(*outDir, "home", ".agent-run")
		if err := seedMarkdownSession(home, "fake-opencode", sessionID); err != nil {
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
				fmt.Fprintf(os.Stderr, "FAIL: port %d busy: %v\n", listenPort, err)
				os.Exit(2)
			}
			fmt.Fprintf(os.Stderr, "port %d busy, using %d\n", *port, p)
			listenPort = p
		}
		baseURL = fmt.Sprintf("http://127.0.0.1:%d", listenPort)
		sessPath = sessionPath

		env := append(os.Environ(), "AGENT_RUN_HOME="+home)
		args := []string{"web", "--port", fmt.Sprint(listenPort), "--no-open", "--agent-runner", "fake-opencode"}
		var webStderr bytes.Buffer
		webCmd = exec.CommandContext(ctx, agentRun, args...)
		webCmd.Env = env
		webCmd.Stdout = os.Stdout
		webCmd.Stderr = io.MultiWriter(os.Stderr, &webStderr)
		fmt.Fprintf(os.Stderr, "starting agent-run web on %s ...\n", baseURL)
		if err := webCmd.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "FAIL: start web: %v\n", err)
			os.Exit(2)
		}
		defer func() {
			if webCmd.Process != nil {
				_ = webCmd.Process.Signal(syscall.SIGTERM)
				_, _ = webCmd.Process.Wait()
			}
		}()

		if err := waitHealth(ctx, baseURL, *token, 45*time.Second); err != nil {
			fmt.Fprintf(os.Stderr, "FAIL: health: %v\nweb stderr:\n%s\n", err, webStderr.String())
			os.Exit(2)
		}
		fmt.Fprintln(os.Stderr, "health: ok")
	}

	probeJS := filepath.Join(repoRoot, "script/debug/markdown-message-card-render/probe.js")
	probeOut := filepath.Join(*outDir, "probe")
	_ = os.MkdirAll(probeOut, 0755)

	report, raw, err := runProbe(ctx, probeJS, baseURL, probeOut, *token, sessPath)
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

	// Stop web before process exit (os.Exit skips defers and would leak agent-run).
	stopWeb := func() {
		if webCmd != nil && webCmd.Process != nil {
			_ = webCmd.Process.Signal(syscall.SIGTERM)
			_, _ = webCmd.Process.Wait()
			webCmd = nil
		}
	}

	switch mode {
	case "repro":
		if report.SymptomPresent {
			fmt.Println("REPRO: markdown not rendered on thinking/response message cards")
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
			os.Exit(1) // non-zero = symptom present (bug-repro inspect RED)
		}
		fmt.Fprintf(os.Stderr, "FAIL: markdown-not-rendered symptom not observed (cannot establish REPRO)\n")
		fmt.Fprintf(os.Stderr, "issues: %v\nokHealthy=%v\n", report.Issues, report.OkHealthy)
		fmt.Fprintf(os.Stderr, "report: %s\n", filepath.Join(*outDir, "probe-report.json"))
		// If already healthy, tip the operator.
		if report.OkHealthy {
			fmt.Fprintln(os.Stderr, "hint: page already looks healthy; use --expect=healthy for VERIFY")
		}
		stopWeb()
		os.Exit(2)

	case "healthy":
		if report.SymptomPresent {
			fmt.Println("FAIL: markdown still not rendered (does not look good)")
			if reason, ok := report.Symptom["reason"].(string); ok {
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
		if !report.OkHealthy {
			fmt.Fprintf(os.Stderr, "FAIL: expected healthy markdown render; issues=%v\n", report.Issues)
			for _, s := range report.Screenshots {
				fmt.Printf("FAIL: screenshot %s\n", s)
			}
			stopWeb()
			os.Exit(1)
		}
		fmt.Println("VERIFY: thinking + response cards render markdown (looks good)")
		for _, s := range report.Screenshots {
			fmt.Printf("VERIFY: screenshot %s\n", s)
		}
		fmt.Printf("VERIFY: report %s\n", filepath.Join(*outDir, "probe-report.json"))
		stopWeb()
		os.Exit(0)
	}
}

func seedMarkdownSession(home, runner, id string) error {
	sessDir := filepath.Join(home, "sessions", id)
	if err := os.MkdirAll(sessDir, 0755); err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	meta := map[string]any{
		"runner":     runner,
		"session_id": id,
		"status":     "idle",
		"created_at": now,
		"updated_at": now,
		"workspace":  "/tmp/demo-workspace",
	}
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(sessDir, "meta.json"), metaBytes, 0644); err != nil {
		return err
	}

	// NDJSON events with think + user + assistant (markdown-rich).
	ts := time.Now().UTC().Add(-3 * time.Minute).UnixMilli()
	lines := []map[string]any{
		{"type": "think", "timestamp": ts, "text": seedThink},
		{"type": "message", "role": "user", "timestamp": ts + 1000, "text": seedUser},
		{"type": "think", "timestamp": ts + 2000, "text": seedThink},
		{
			"type":      "message",
			"role":      "assistant",
			"timestamp": ts + 3000,
			"text":      seedAssistant,
		},
	}
	var b strings.Builder
	for _, ev := range lines {
		raw, err := json.Marshal(ev)
		if err != nil {
			return err
		}
		b.Write(raw)
		b.WriteByte('\n')
	}
	return os.WriteFile(filepath.Join(sessDir, "events.jsonl"), []byte(b.String()), 0644)
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
