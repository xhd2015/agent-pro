// message-card-reorder-followup: reproduce / verify live message-card reorder
// on grok-tty follow-up (second user bubble causes first user card to jump).
//
// Modes:
//
//	(default) repro  — expect live DOM reorder symptom after second send.
//	                   Exit 1 + REPRO: lines when symptom present (loop RED).
//	--expect=healthy — expect live timeline keeps correct interleaving
//	                   (assistant between the two user cards) and reload healthy.
//	                   Exit 0 only when green (VERIFY gate after fix).
//
// Usage (from repo root):
//
//	go run ./script/debug/message-card-reorder-followup
//	go run ./script/debug/message-card-reorder-followup --expect=healthy
//	go run ./script/debug/message-card-reorder-followup -port=0 -out=/tmp/reorder-loop
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
)

type probeReport struct {
	BaseURL        string         `json:"baseURL"`
	SymptomPresent bool           `json:"symptomPresent"`
	Symptom        map[string]any `json:"symptom"`
	OkHealthy      bool           `json:"okHealthy"`
	Issues         []string       `json:"issues"`
	Page           map[string]any `json:"page"`
	Snapshots      map[string]any `json:"snapshots"`
	Screenshots    []string       `json:"screenshots"`
}

func main() {
	expect := flag.String("expect", "repro", "repro | healthy")
	port := flag.Int("port", 0, "web listen port (0 = free ephemeral)")
	outDir := flag.String("out", "", "output dir (default /tmp/message-card-reorder-followup-<ts>)")
	skipBuild := flag.Bool("skip-build", false, "reuse binaries under -out/bin if present")
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
		*outDir = filepath.Join(os.TempDir(), fmt.Sprintf("message-card-reorder-followup-%d", time.Now().Unix()))
	}
	if err := os.MkdirAll(*outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: mkdir out: %v\n", err)
		os.Exit(2)
	}
	fmt.Fprintf(os.Stderr, "message-card-reorder-followup: out=%s expect=%s\n", *outDir, mode)

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
	mockGrok := filepath.Join(binDir, "llm-mock-run-grok")

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
		fmt.Fprintln(os.Stderr, "building llm-mock-run-grok...")
		if err := goBuild(repoRoot, "./agent/llm/llm-mock/llm-mock-run-grok", mockGrok); err != nil {
			fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
			os.Exit(2)
		}
	} else {
		for _, p := range []string{agentRun, mockGrok} {
			if _, err := os.Stat(p); err != nil {
				fmt.Fprintf(os.Stderr, "FAIL: missing binary %s (drop -skip-build)\n", p)
				os.Exit(2)
			}
		}
	}

	// Fresh homes every run so leftover sessions never pollute REPRO.
	runStamp := fmt.Sprintf("run-%d", time.Now().UnixNano())
	home := filepath.Join(*outDir, "homes", runStamp, ".agent-run")
	grokHome := filepath.Join(*outDir, "homes", runStamp, "grok-home")
	if err := os.MkdirAll(home, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: mkdir home: %v\n", err)
		os.Exit(2)
	}
	if err := os.MkdirAll(grokHome, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: mkdir grok-home: %v\n", err)
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
	} else {
		// If busy, pick free nearby
		if !portFree(listenPort) {
			p, err := freePort()
			if err != nil {
				fmt.Fprintf(os.Stderr, "FAIL: port %d busy and no free port: %v\n", listenPort, err)
				os.Exit(2)
			}
			fmt.Fprintf(os.Stderr, "port %d busy, using %d\n", *port, p)
			listenPort = p
		}
	}
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", listenPort)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hook := multiTurnMockGrokHook()
	env := append(os.Environ(),
		"AGENT_RUN_HOME="+home,
		"PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"LLM_MOCK_RUN_GROK_COMMAND="+hook,
		"AGENT_RUN_GROK_TTY_GROK_SESSION_ID=a1111111-1111-4111-8111-111111111111",
	)
	args := []string{
		"web",
		"--port", fmt.Sprint(listenPort),
		"--no-open",
		"--agent-runner", "grok-tty",
		"--grok-home", grokHome,
		"--grok-tty-runner-binary", mockGrok,
	}
	// open API (no --token) so browser needs no auth
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
	defer func() {
		if webCmd.Process != nil {
			_ = webCmd.Process.Signal(syscall.SIGTERM)
			_, _ = webCmd.Process.Wait()
		}
	}()

	if err := waitHealth(ctx, baseURL, "", 45*time.Second); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: health: %v\nweb stderr:\n%s\n", err, webStderr.String())
		os.Exit(2)
	}
	fmt.Fprintln(os.Stderr, "health: ok")

	probeJS := filepath.Join(repoRoot, "script/debug/message-card-reorder-followup/probe.js")
	probeOut := filepath.Join(*outDir, "probe")
	_ = os.MkdirAll(probeOut, 0755)

	report, raw, err := runProbe(ctx, probeJS, baseURL, probeOut, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: playwright probe: %v\nraw:\n%s\n", err, raw)
		os.Exit(2)
	}
	_ = os.WriteFile(filepath.Join(*outDir, "probe-report.json"), []byte(mustJSON(report)), 0644)

	fmt.Fprintf(os.Stderr, "probe: symptomPresent=%v okHealthy=%v issues=%v\n",
		report.SymptomPresent, report.OkHealthy, report.Issues)
	if report.Symptom != nil {
		fmt.Fprintf(os.Stderr, "probe symptom: %v\n", report.Symptom["reason"])
	}

	switch mode {
	case "repro":
		if report.SymptomPresent {
			fmt.Println("REPRO: live message-card reorder on grok-tty follow-up")
			if reason, ok := report.Symptom["reason"].(string); ok && reason != "" {
				fmt.Printf("REPRO: %s\n", reason)
			}
			if users := snapshotUsers(report, "liveAfterSecond"); len(users) > 0 {
				fmt.Printf("REPRO: live users=%v\n", users)
			}
			if users := snapshotUsers(report, "afterReload"); len(users) > 0 {
				fmt.Printf("REPRO: after reload users=%v (okHealthy=%v)\n", users, report.OkHealthy)
			}
			for _, s := range report.Screenshots {
				fmt.Printf("REPRO: screenshot %s\n", s)
			}
			fmt.Printf("REPRO: report %s\n", filepath.Join(*outDir, "probe-report.json"))
			os.Exit(1) // non-zero = symptom present (bug-repro inspect RED)
		}
		fmt.Fprintf(os.Stderr, "FAIL: reorder symptom not observed (cannot establish REPRO)\n")
		fmt.Fprintf(os.Stderr, "issues: %v\n", report.Issues)
		fmt.Fprintf(os.Stderr, "report: %s\n", filepath.Join(*outDir, "probe-report.json"))
		os.Exit(2)

	case "healthy":
		if report.SymptomPresent {
			fmt.Println("FAIL: reorder symptom still present after fix attempt")
			if reason, ok := report.Symptom["reason"].(string); ok {
				fmt.Printf("FAIL: %s\n", reason)
			}
			os.Exit(1)
		}
		if !report.OkHealthy {
			fmt.Fprintf(os.Stderr, "FAIL: expected healthy timeline after reload; issues=%v\n", report.Issues)
			os.Exit(1)
		}
		// Also require live path healthy (not only reload).
		if liveHit := liveSymptomFromReport(report); liveHit {
			fmt.Println("FAIL: live path still reorders (reload-only green is insufficient)")
			os.Exit(1)
		}
		fmt.Println("VERIFY: follow-up keeps correct message order (live + reload)")
		fmt.Printf("VERIFY: report %s\n", filepath.Join(*outDir, "probe-report.json"))
		os.Exit(0)
	}
}

func liveSymptomFromReport(r probeReport) bool {
	return r.SymptomPresent
}

func snapshotUsers(r probeReport, key string) []string {
	if r.Snapshots == nil {
		return nil
	}
	raw, ok := r.Snapshots[key]
	if !ok {
		return nil
	}
	m, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	u, ok := m["users"]
	if !ok {
		return nil
	}
	arr, ok := u.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, v := range arr {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func runProbe(ctx context.Context, probeJS, baseURL, outDir, token string) (probeReport, string, error) {
	var zero probeReport
	cmd := exec.CommandContext(ctx, "playwright-debug", "run", probeJS, baseURL, outDir, token)
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

// multiTurnMockGrokHook handles one turn per process (each agent-run spawn).
// Same shape as web-layout llmMockGrokLayoutHook so discovery finds updates.jsonl.
// Truncates updates per turn; agent-run already persisted prior turns in events.jsonl.
func multiTurnMockGrokHook() string {
	// Fixed progressive assistant text (no nested python quoting).
	// Includes both prompt substrings so UI wait-for-assistant is stable.
	const part1 = "MOCK_REPLY:"
	const part2 = "MOCK_REPLY: done for "
	const part3 = "MOCK_REPLY: done for turn"
	// Match web-layout llmMockGrokLayoutHook: banner → read → write updates.
	// Do NOT pre-create empty updates.jsonl (discovery can hang on empty file).
	return fmt.Sprintf(`sh -c '
printf "GROK_TTY_BANNER\nGrok › "
read -r line || true
submitted="${line:-run ls}"
wd=$(pwd)
enc=$(python3 -c '"'"'import os,sys,urllib.parse
p=os.path.abspath(sys.argv[1])
if p.startswith("/private/var/"): p="/var/"+p[len("/private/var/"):]
elif p.startswith("/private/tmp/"): p="/tmp/"+p[len("/private/tmp/"):]
print(urllib.parse.quote(p, safe=""))'"'"' "$wd")
dir="$GROK_HOME/sessions/$enc/a1111111-1111-4111-8111-111111111111"
mkdir -p "$dir"
now=$(date -u +%%Y-%%m-%%dT%%H:%%M:%%SZ)
cat > "$dir/summary.json" <<EOF
{"info":{"cwd":"$wd","sessionId":"a1111111-1111-4111-8111-111111111111","openedAt":"$now"},"created_at":"$now"}
EOF
updates="$dir/updates.jsonl"
printf %%s\\n "{\"sessionUpdate\":\"user_message_chunk\",\"content\":{\"type\":\"text\",\"text\":\"$submitted\"}}" > "$updates"
stream_chunk() { printf %%s\\n "$1" >> "$updates"; }
stream_chunk "{\"sessionUpdate\":\"agent_message_chunk\",\"content\":{\"type\":\"text\",\"text\":\"%s\"}}"
sleep 0.35
stream_chunk "{\"sessionUpdate\":\"agent_message_chunk\",\"content\":{\"type\":\"text\",\"text\":\"%s\"}}"
sleep 0.35
stream_chunk "{\"sessionUpdate\":\"agent_message_chunk\",\"content\":{\"type\":\"text\",\"text\":\"%s\"}}"
stream_chunk "{\"sessionUpdate\":\"turn_completed\"}"
sleep 1
exit 0
'`, part1, part2, part3)
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
