// agent-run-web-loop demos an automated inspect → report loop for agent-run web.
//
// Flow per run:
//   1. optional: go run ./script/install
//   2. build agent-run (+ optional llm-mock-run-grok)
//   3. start agent-run web on :8192 (or reuse if already healthy)
//   4. loop: health check → playwright-debug probe → screenshot/status report
//
// Usage (from repo root):
//
//	go run ./script/debug/agent-run-web-loop
//	go run ./script/debug/agent-run-web-loop -install -loops=3
//	go run ./script/debug/agent-run-web-loop -reuse-web -loops=5 -out=/tmp/web-loop
//	go run ./script/debug/agent-run-web-loop -no-mock   # real grok on PATH
//
// Requires playwright-debug on PATH.
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
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type probeReport struct {
	Iteration   int               `json:"iteration"`
	BaseURL     string            `json:"baseURL"`
	Prompt      string            `json:"prompt"`
	Viewport    map[string]int    `json:"viewport,omitempty"`
	Screenshots []string          `json:"screenshots"`
	Page        map[string]string `json:"page"`
	Elements    map[string]any    `json:"elements"`
	Layout      map[string]any    `json:"layout,omitempty"`
	Network     struct {
		SessionDetailGets int `json:"sessionDetailGets"`
		SSEStreams        int `json:"sseStreams"`
	} `json:"network"`
	Issues []string `json:"issues"`
	OK     bool     `json:"ok"`
}

type loopSummary struct {
	StartedAt  string        `json:"started_at"`
	FinishedAt string        `json:"finished_at"`
	BaseURL    string        `json:"base_url"`
	OutDir     string        `json:"out_dir"`
	Loops      int           `json:"loops"`
	Iterations []probeReport `json:"iterations"`
	AllOK      bool          `json:"all_ok"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "agent-run-web-loop: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		doInstall = flag.Bool("install", false, "run go run ./script/install first")
		skipBuild = flag.Bool("skip-build", false, "skip building agent-run binary")
		loops     = flag.Int("loops", 3, "number of probe iterations")
		port      = flag.Int("port", 8192, "agent-run web port")
		outDir    = flag.String("out", "", "output dir for screenshots and JSON reports (default: /tmp/agent-run-web-loop-<ts>)")
		prompt    = flag.String("prompt", "hello from web loop probe", "composer message to send each iteration")
		useMock   = flag.Bool("mock", true, "use llm-mock-run-grok for reproducible grok-tty runs")
		reuseWeb  = flag.Bool("reuse-web", false, "skip starting web if health check already passes on port")
		webToken  = flag.String("token", "", "agent-run web API token (empty=open API; auto=generate+require auth)")
		pause     = flag.Duration("pause", 2*time.Second, "pause between loop iterations")
	)
	flag.Parse()

	repoRoot, err := findRepoRoot()
	if err != nil {
		return err
	}

	if *outDir == "" {
		*outDir = filepath.Join(os.TempDir(), fmt.Sprintf("agent-run-web-loop-%d", time.Now().Unix()))
	}
	if err := os.MkdirAll(*outDir, 0755); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "agent-run-web-loop: out=%s\n", *outDir)

	if *doInstall {
		fmt.Fprintln(os.Stderr, "running script/install...")
		install := exec.Command("go", "run", "./script/install")
		install.Dir = repoRoot
		install.Stdout = os.Stdout
		install.Stderr = os.Stderr
		if err := install.Run(); err != nil {
			return fmt.Errorf("script/install: %w", err)
		}
	}
	if err := buildAgentRunFrontend(repoRoot); err != nil {
		return err
	}

	binDir := filepath.Join(*outDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}
	agentRun := filepath.Join(binDir, "agent-run")
	if !*skipBuild {
		fmt.Fprintln(os.Stderr, "building agent-run...")
		if err := goBuild(repoRoot, "./cmd/agent-run", agentRun); err != nil {
			return err
		}
	} else if _, err := os.Stat(agentRun); err != nil {
		return fmt.Errorf("agent-run binary missing at %s (drop -skip-build)", agentRun)
	}

	home := filepath.Join(*outDir, ".agent-run")
	grokHome := filepath.Join(*outDir, "grok-home")
	if err := os.MkdirAll(home, 0755); err != nil {
		return err
	}

	probeJS := filepath.Join(repoRoot, "script/debug/agent-run-web-loop/probe.js")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	listenPort := *port
	var webCmd *exec.Cmd
	var bearer string
	startedWeb := false
	if *reuseWeb {
		baseURL := fmt.Sprintf("http://127.0.0.1:%d", listenPort)
		if err := waitHealth(ctx, baseURL, *webToken, 2*time.Second); err == nil {
			fmt.Fprintf(os.Stderr, "reusing existing web server at %s\n", baseURL)
			bearer = *webToken
		} else {
			*reuseWeb = false
		}
	}
	if !*reuseWeb {
		free, err := findFreePort(listenPort, listenPort+99)
		if err != nil {
			return err
		}
		if free != listenPort {
			fmt.Fprintf(os.Stderr, "port %d busy, using %d\n", listenPort, free)
			listenPort = free
		}
	}
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", listenPort)

	if !*reuseWeb {
		env := append(os.Environ(),
			"AGENT_RUN_HOME="+home,
			"PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"),
		)
		args := []string{"web", "--port", fmt.Sprint(listenPort), "--no-open"}
		if tok := strings.TrimSpace(*webToken); tok != "" {
			args = append(args, "--token", tok)
		}
		var webStderr bytes.Buffer
		if *useMock {
			mockBin := filepath.Join(binDir, "llm-mock-run-grok")
			if err := os.MkdirAll(grokHome, 0755); err != nil {
				return err
			}
			fmt.Fprintln(os.Stderr, "building llm-mock-run-grok...")
			if err := goBuild(repoRoot, "./agent/llm/llm-mock/llm-mock-run-grok", mockBin); err != nil {
				return err
			}
			args = append(args,
				"--agent-runner", "grok-tty",
				"--grok-home", grokHome,
				"--grok-tty-runner-binary", mockBin,
			)
			env = append(env,
				"LLM_MOCK_RUN_GROK_COMMAND="+mockGrokHook(*prompt),
				"AGENT_RUN_GROK_TTY_GROK_SESSION_ID=a1111111-1111-4111-8111-111111111111",
			)
		}
		fmt.Fprintf(os.Stderr, "starting agent-run web on %s ...\n", baseURL)
		webCmd = exec.CommandContext(ctx, agentRun, args...)
		webCmd.Env = env
		webCmd.Stdout = os.Stdout
		webCmd.Stderr = io.MultiWriter(os.Stderr, &webStderr)
		if err := webCmd.Start(); err != nil {
			return fmt.Errorf("start agent-run web: %w", err)
		}
		startedWeb = true
		defer func() {
			if webCmd != nil && webCmd.Process != nil {
				_ = webCmd.Process.Signal(syscall.SIGTERM)
				_, _ = webCmd.Process.Wait()
			}
		}()
		bearer = strings.TrimSpace(*webToken)
		if err := waitWebReady(ctx, webCmd, baseURL, bearer, &webStderr, 45*time.Second); err != nil {
			return err
		}
		if parsed := parseWebTokenFromStderr(webStderr.String()); parsed != "" {
			bearer = parsed
		}
	}

	defaultRunner, err := fetchDefaultRunner(ctx, baseURL, bearer)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "default runner: %s\n", defaultRunner)

	if _, err := exec.LookPath("playwright-debug"); err != nil {
		return fmt.Errorf("playwright-debug not on PATH: %w", err)
	}

	summary := loopSummary{
		StartedAt: time.Now().UTC().Format(time.RFC3339),
		BaseURL:   baseURL,
		OutDir:    *outDir,
		Loops:     *loops,
		AllOK:     true,
	}

	for i := 1; i <= *loops; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		fmt.Fprintf(os.Stderr, "\n=== loop iteration %d/%d ===\n", i, *loops)

		if err := waitHealth(ctx, baseURL, bearer, 5*time.Second); err != nil {
			return fmt.Errorf("iteration %d health: %w", i, err)
		}
		fmt.Fprintln(os.Stderr, "health: ok")

		iterDir := filepath.Join(*outDir, fmt.Sprintf("iter-%d", i))
		if err := os.MkdirAll(iterDir, 0755); err != nil {
			return err
		}

		report, err := runPlaywrightProbe(ctx, probeJS, baseURL, iterDir, i, *prompt, bearer, defaultRunner)
		if err != nil {
			return fmt.Errorf("iteration %d playwright: %w", i, err)
		}
		summary.Iterations = append(summary.Iterations, report)
		if !report.OK {
			summary.AllOK = false
		}

		reportPath := filepath.Join(iterDir, "report.json")
		if err := writeJSON(reportPath, report); err != nil {
			return err
		}
		printHumanReport(i, report, reportPath)

		if i < *loops {
			time.Sleep(*pause)
		}
	}

	summary.FinishedAt = time.Now().UTC().Format(time.RFC3339)
	summaryPath := filepath.Join(*outDir, "loop-summary.json")
	if err := writeJSON(summaryPath, summary); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "\nloop complete: all_ok=%v summary=%s\n", summary.AllOK, summaryPath)
	if startedWeb {
		fmt.Fprintln(os.Stderr, "web server stopping...")
	}
	if !summary.AllOK {
		return fmt.Errorf("one or more iterations reported issues (see %s)", summaryPath)
	}
	return nil
}

func runPlaywrightProbe(ctx context.Context, probeJS, baseURL, outDir string, iter int, prompt, bearer, runner string) (probeReport, error) {
	var zero probeReport
	cmd := exec.CommandContext(ctx, "playwright-debug", "run", probeJS, baseURL, outDir, fmt.Sprint(iter), prompt, bearer, runner)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)

	if err := cmd.Run(); err != nil {
		return zero, fmt.Errorf("%w\nstderr: %s", err, stderr.String())
	}

	line := extractReportJSON(stdout.String())
	if line == "" {
		return zero, fmt.Errorf("no REPORT_JSON in playwright output\nstdout: %s", stdout.String())
	}
	var report probeReport
	if err := json.Unmarshal([]byte(line), &report); err != nil {
		return zero, fmt.Errorf("parse report: %w", err)
	}
	return report, nil
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

func printHumanReport(iter int, r probeReport, path string) {
	status := "OK"
	if !r.OK {
		status = "ISSUES"
	}
	fmt.Fprintf(os.Stderr, "iteration %d: %s\n", iter, status)
	fmt.Fprintf(os.Stderr, "  url: %s\n", r.Page["urlAfterSend"])
	if u, ok := r.Elements["userCount"].(float64); ok {
		fmt.Fprintf(os.Stderr, "  user messages: %.0f\n", u)
	}
	if a, ok := r.Elements["assistantCount"].(float64); ok {
		fmt.Fprintf(os.Stderr, "  assistant messages: %.0f\n", a)
	}
	fmt.Fprintf(os.Stderr, "  sse streams: %d  detail GETs: %d\n", r.Network.SSEStreams, r.Network.SessionDetailGets)
	for _, s := range r.Screenshots {
		fmt.Fprintf(os.Stderr, "  screenshot: %s\n", s)
	}
	for _, issue := range r.Issues {
		fmt.Fprintf(os.Stderr, "  issue: %s\n", issue)
	}
	fmt.Fprintf(os.Stderr, "  report: %s\n", path)
}

func findFreePort(start, end int) (int, error) {
	for p := start; p <= end; p++ {
		l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", p))
		if err == nil {
			_ = l.Close()
			return p, nil
		}
	}
	return 0, fmt.Errorf("no free TCP port in %d-%d", start, end)
}

func fetchDefaultRunner(ctx context.Context, baseURL, bearer string) (string, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	url := strings.TrimRight(baseURL, "/") + "/api/agent-run/runners"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch runners: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch runners: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var parsed struct {
		Default string `json:"default"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("parse runners: %w", err)
	}
	if strings.TrimSpace(parsed.Default) == "" {
		return "opencode", nil
	}
	return parsed.Default, nil
}

func parseWebTokenFromStderr(stderr string) string {
	for _, line := range strings.Split(stderr, "\n") {
		line = strings.TrimSpace(line)
		const prefix = "agent-run web token: "
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix))
		}
	}
	return ""
}

func waitWebReady(ctx context.Context, cmd *exec.Cmd, baseURL, bearer string, stderr *bytes.Buffer, timeout time.Duration) error {
	if cmd == nil {
		return waitHealth(ctx, baseURL, bearer, timeout)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-done:
			if err != nil {
				return fmt.Errorf("agent-run web exited before healthy: %w", err)
			}
			return fmt.Errorf("agent-run web exited before healthy")
		default:
		}
		tok := bearer
		if stderr != nil {
			if parsed := parseWebTokenFromStderr(stderr.String()); parsed != "" {
				tok = parsed
			}
		}
		if err := waitHealth(ctx, baseURL, tok, 2*time.Second); err == nil {
			return nil
		}
		time.Sleep(150 * time.Millisecond)
	}
	return fmt.Errorf("agent-run web not healthy within %s: %s", timeout, strings.TrimRight(baseURL, "/")+"/api/agent-run/health")
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

func buildAgentRunFrontend(repoRoot string) error {
	fmt.Fprintln(os.Stderr, "building frontend-agent-run...")
	cmd := exec.Command("bun", "run", "build")
	cmd.Dir = filepath.Join(repoRoot, "frontend-agent-run")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("frontend-agent-run build: %w\n%s", err, string(out))
	}
	return nil
}

func goBuild(repoRoot, pkg, out string) error {
	cmd := exec.Command("go", "build", "-buildvcs=false", "-o", out, pkg)
	cmd.Dir = repoRoot
	if outBytes, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go build %s: %w\n%s", pkg, err, string(outBytes))
	}
	return nil
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

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// mockGrokHook is a minimal grok-tty chrome hook (same idea as web-layout tests).
func mockGrokHook(prompt string) string {
	marker := "WEB_LOOP_PROBE_REPLY:" + prompt
	n := len(marker)
	third := n / 3
	twoThird := (2 * n) / 3
	if third < 1 {
		third = 1
	}
	if twoThird <= third {
		twoThird = third + 1
	}
	if twoThird >= n {
		twoThird = n - 1
	}
	part1, part2, part3 := marker[:third], marker[:twoThird], marker
	return fmt.Sprintf(`sh -c '
printf "GROK_TTY_BANNER\nGrok › "
read -r line || true
submitted="${line:-%s}"
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
sleep 0.3
stream_chunk "{\"sessionUpdate\":\"agent_message_chunk\",\"content\":{\"type\":\"text\",\"text\":\"%s\"}}"
sleep 0.3
stream_chunk "{\"sessionUpdate\":\"agent_message_chunk\",\"content\":{\"type\":\"text\",\"text\":\"%s\"}}"
printf %%s\\n "{\"sessionUpdate\":\"turn_completed\"}" >> "$updates"
sleep 1
exit 0
'`, prompt, part1, part2, part3)
}

