// session-card-id-footer-capped: reproduce / verify home session card bottom
// footer showing mid-ellipsis shortSessionId under a prompt title.
//
// Modes:
//
//	(default) repro  — expect capped id footer present (exit 1 + REPRO:)
//	--expect=healthy — expect no mid-ellipsis id footer on prompt cards (exit 0)
//
// Usage (repo root):
//
//	go run ./script/debug/session-card-id-footer-capped
//	go run ./script/debug/session-card-id-footer-capped --expect=healthy
//	go run ./script/debug/session-card-id-footer-capped -url=http://127.0.0.1:8821
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
	// Long human-readable session id (user screenshot style).
	seedSessionID = "brainstorm-add-terminal-color-skill-20260713-083040"
	seedPrompt    = "/brainstorm add terminal color skill"
	runnerID      = "fake-codex"
)

type probeReport struct {
	BaseURL        string         `json:"baseURL"`
	SymptomPresent bool           `json:"symptomPresent"`
	OkHealthy      bool           `json:"okHealthy"`
	Reasons        []string       `json:"reasons"`
	Issues         []string       `json:"issues"`
	Symptom        map[string]any `json:"symptom"`
	Snapshots      map[string]any `json:"snapshots"`
	Screenshots    []string       `json:"screenshots"`
}

func main() {
	expect := flag.String("expect", "repro", "repro | healthy")
	port := flag.Int("port", 0, "web listen port (0 = free ephemeral)")
	outDir := flag.String("out", "", "output dir")
	skipBuild := flag.Bool("skip-build", false, "reuse -out/bin/agent-run if present")
	token := flag.String("token", "", "optional API token")
	baseURLFlag := flag.String("url", "", "probe this URL only (skip seed/build)")
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
		*outDir = filepath.Join(os.TempDir(), fmt.Sprintf("session-card-id-footer-capped-%d", time.Now().Unix()))
	}
	if err := os.MkdirAll(*outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: mkdir: %v\n", err)
		os.Exit(2)
	}
	fmt.Fprintf(os.Stderr, "session-card-id-footer-capped: out=%s expect=%s\n", *outDir, mode)

	if _, err := exec.LookPath("playwright-debug"); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: playwright-debug not on PATH: %v\n", err)
		os.Exit(2)
	}

	var (
		baseURL string
		stopWeb func()
	)

	if strings.TrimSpace(*baseURLFlag) != "" {
		baseURL = strings.TrimRight(strings.TrimSpace(*baseURLFlag), "/")
		fmt.Fprintf(os.Stderr, "using external URL %s\n", baseURL)
	} else {
		if _, err := exec.LookPath("bun"); err != nil {
			fmt.Fprintf(os.Stderr, "FAIL: bun not on PATH: %v\n", err)
			os.Exit(2)
		}
		binDir := filepath.Join(*outDir, "bin")
		_ = os.MkdirAll(binDir, 0755)
		agentRun := filepath.Join(binDir, "agent-run")
		if !*skipBuild {
			fmt.Fprintln(os.Stderr, "building frontend-agent-run...")
			cmd := exec.Command("bun", "run", "build")
			cmd.Dir = filepath.Join(repoRoot, "frontend-agent-run")
			if out, err := cmd.CombinedOutput(); err != nil {
				fmt.Fprintf(os.Stderr, "FAIL: frontend build: %v\n%s\n", err, out)
				os.Exit(2)
			}
			fmt.Fprintln(os.Stderr, "building agent-run...")
			if err := goBuild(repoRoot, "./cmd/agent-run", agentRun); err != nil {
				fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
				os.Exit(2)
			}
		} else if _, err := os.Stat(agentRun); err != nil {
			fmt.Fprintf(os.Stderr, "FAIL: missing %s\n", agentRun)
			os.Exit(2)
		}

		home := filepath.Join(*outDir, "home", ".agent-run")
		if err := seedSession(home); err != nil {
			fmt.Fprintf(os.Stderr, "FAIL: seed: %v\n", err)
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
			listenPort = p
		}
		baseURL = fmt.Sprintf("http://127.0.0.1:%d", listenPort)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		env := append(os.Environ(), "AGENT_RUN_HOME="+home)
		webCmd := exec.CommandContext(ctx, agentRun, "web", "--port", fmt.Sprint(listenPort), "--no-open", "--agent-runner", runnerID)
		webCmd.Env = env
		var webStderr bytes.Buffer
		webCmd.Stdout = os.Stdout
		webCmd.Stderr = io.MultiWriter(os.Stderr, &webStderr)
		if err := webCmd.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "FAIL: start web: %v\n", err)
			os.Exit(2)
		}
		stopWeb = func() {
			if webCmd.Process != nil {
				_ = webCmd.Process.Signal(syscall.SIGTERM)
				_, _ = webCmd.Process.Wait()
			}
		}
		defer stopWeb()

		if err := waitHealth(ctx, baseURL, *token, 45*time.Second); err != nil {
			fmt.Fprintf(os.Stderr, "FAIL: health: %v\n%s\n", err, webStderr.String())
			os.Exit(2)
		}
		fmt.Fprintln(os.Stderr, "health: ok")
	}

	probeJS := filepath.Join(repoRoot, "script/debug/session-card-id-footer-capped/probe.js")
	probeOut := filepath.Join(*outDir, "probe")
	_ = os.MkdirAll(probeOut, 0755)

	report, raw, err := runProbe(context.Background(), probeJS, baseURL, probeOut, *token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: playwright: %v\n%s\n", err, raw)
		os.Exit(2)
	}
	_ = os.WriteFile(filepath.Join(*outDir, "probe-report.json"), []byte(mustJSON(report)), 0644)

	fmt.Fprintf(os.Stderr, "probe: symptomPresent=%v okHealthy=%v reasons=%v\n",
		report.SymptomPresent, report.OkHealthy, report.Reasons)

	switch mode {
	case "repro":
		if report.SymptomPresent {
			fmt.Println("REPRO: session card bottom shows capped mid-ellipsis session id under prompt")
			for _, r := range report.Reasons {
				fmt.Printf("REPRO: %s\n", r)
			}
			if report.Symptom != nil {
				fmt.Printf("REPRO: symptom %s\n", compactJSON(report.Symptom))
			}
			for _, s := range report.Screenshots {
				fmt.Printf("REPRO: screenshot %s\n", s)
			}
			fmt.Printf("REPRO: report %s\n", filepath.Join(*outDir, "probe-report.json"))
			if stopWeb != nil {
				stopWeb()
			}
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "FAIL: capped id footer symptom not observed\n")
		if report.OkHealthy {
			fmt.Fprintln(os.Stderr, "hint: already healthy; use --expect=healthy")
		}
		if stopWeb != nil {
			stopWeb()
		}
		os.Exit(2)

	case "healthy":
		if report.SymptomPresent || !report.OkHealthy {
			fmt.Println("FAIL: session card id footer still capped / not healthy")
			for _, r := range report.Reasons {
				fmt.Printf("FAIL: %s\n", r)
			}
			if stopWeb != nil {
				stopWeb()
			}
			os.Exit(1)
		}
		fmt.Println("VERIFY: no mid-ellipsis session-id footer on prompt cards")
		for _, s := range report.Screenshots {
			fmt.Printf("VERIFY: screenshot %s\n", s)
		}
		fmt.Printf("VERIFY: report %s\n", filepath.Join(*outDir, "probe-report.json"))
		if stopWeb != nil {
			stopWeb()
		}
		os.Exit(0)
	}
}

func seedSession(home string) error {
	if err := os.MkdirAll(filepath.Join(home, "sessions"), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(home, "sessions", ".layout"), []byte(`{"version":2}`+"\n"), 0644); err != nil {
		return err
	}
	sessDir := filepath.Join(home, "sessions", seedSessionID)
	if err := os.MkdirAll(sessDir, 0755); err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	meta := map[string]any{
		"runner":         runnerID,
		"session_id":     seedSessionID,
		"initial_prompt": seedPrompt,
		"status":         "running",
		"workspace":      filepath.Join(home, "workspace"),
		"created_at":     now,
		"updated_at":     now,
	}
	b, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(sessDir, "meta.json"), b, 0644); err != nil {
		return err
	}
	events := fmt.Sprintf(
		`{"type":"message","role":"user","text":%q,"timestamp":1}`+"\n",
		seedPrompt,
	)
	return os.WriteFile(filepath.Join(sessDir, "events.jsonl"), []byte(events), 0644)
}

func runProbe(ctx context.Context, probeJS, baseURL, outDir, token string) (probeReport, string, error) {
	var zero probeReport
	cmd := exec.CommandContext(ctx, "playwright-debug", "run", probeJS, baseURL, outDir, token)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	if err := cmd.Run(); err != nil {
		return zero, stdout.String() + "\n" + stderr.String(), err
	}
	line := extractReportJSON(stdout.String())
	if line == "" {
		return zero, stdout.String(), fmt.Errorf("no REPORT_JSON")
	}
	var report probeReport
	if err := json.Unmarshal([]byte(line), &report); err != nil {
		return zero, line, err
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
	return fmt.Errorf("health timeout: %s", url)
}

func goBuild(repoRoot, pkg, out string) error {
	cmd := exec.Command("go", "build", "-buildvcs=false", "-o", out, pkg)
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
			return "", fmt.Errorf("go.mod not found")
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

func compactJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}
