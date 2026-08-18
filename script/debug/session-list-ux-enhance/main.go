// session-list-ux-enhance: reproduce / verify session list home UX:
//   1) Load more only at end of list (inside scroll content)
//   2) No auto-load on scroll (explicit button click)
//   3) Enter session → back preserves list scroll (lift state to App)
//
// Modes:
//
//	(default) repro  — expect at least one of the three symptoms present.
//	                   Exit 1 + REPRO: lines when symptom present (loop RED).
//	--expect=healthy — expect all three fixed.
//	                   Exit 0 only when green (VERIFY gate after fix).
//
// Usage (from repo root):
//
//	go run ./script/debug/session-list-ux-enhance
//	go run ./script/debug/session-list-ux-enhance --expect=healthy
//	go run ./script/debug/session-list-ux-enhance -out=/tmp/sl-ux
//	go run ./script/debug/session-list-ux-enhance -url=http://127.0.0.1:8821
//
// Requires: go, bun, playwright-debug on PATH (unless -url only + -skip-build).
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
	// More than frontend PAGE_SIZE (30) so has_more + scroll + load-more apply.
	seedSessionCount = 55
	runnerID         = "fake-codex"
)

type probeReport struct {
	BaseURL        string         `json:"baseURL"`
	SymptomPresent bool           `json:"symptomPresent"`
	OkHealthy      bool           `json:"okHealthy"`
	Issues         []string       `json:"issues"`
	Symptom        map[string]any `json:"symptom"`
	Reasons        []string       `json:"reasons"`
	Snapshots      map[string]any `json:"snapshots"`
	Screenshots    []string       `json:"screenshots"`
	LoadMoreClick  map[string]any `json:"loadMoreClick"`
	Page           map[string]any `json:"page"`
}

func main() {
	expect := flag.String("expect", "repro", "repro | healthy")
	port := flag.Int("port", 0, "web listen port (0 = free ephemeral)")
	outDir := flag.String("out", "", "output dir (default /tmp/session-list-ux-enhance-<ts>)")
	skipBuild := flag.Bool("skip-build", false, "reuse binaries under -out/bin if present")
	token := flag.String("token", "", "optional API token")
	baseURLFlag := flag.String("url", "", "if set, probe this URL only (skip local web seed/build)")
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
		*outDir = filepath.Join(os.TempDir(), fmt.Sprintf("session-list-ux-enhance-%d", time.Now().Unix()))
	}
	if err := os.MkdirAll(*outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: mkdir out: %v\n", err)
		os.Exit(2)
	}
	fmt.Fprintf(os.Stderr, "session-list-ux-enhance: out=%s expect=%s\n", *outDir, mode)

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
		fmt.Fprintf(os.Stderr, "using external URL %s (no seed)\n", baseURL)
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
			if err := goBuild(repoRoot, "./agent-run", agentRun); err != nil {
				fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
				os.Exit(2)
			}
		} else if _, err := os.Stat(agentRun); err != nil {
			fmt.Fprintf(os.Stderr, "FAIL: missing binary %s (drop -skip-build)\n", agentRun)
			os.Exit(2)
		}

		home := filepath.Join(*outDir, "home", ".agent-run")
		if err := seedManySessions(home, runnerID, seedSessionCount); err != nil {
			fmt.Fprintf(os.Stderr, "FAIL: seed sessions: %v\n", err)
			os.Exit(2)
		}
		fmt.Fprintf(os.Stderr, "seeded %d sessions under %s\n", seedSessionCount, home)

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
		baseURL = fmt.Sprintf("http://127.0.0.1:%d", listenPort)

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
		stopWeb = func() {
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
		// Confirm pagination surface.
		if err := waitHasMore(ctx, baseURL, *token, 15*time.Second); err != nil {
			fmt.Fprintf(os.Stderr, "FAIL: sessions pagination: %v\n", err)
			os.Exit(2)
		}
		fmt.Fprintln(os.Stderr, "health + has_more: ok")
	}

	probeJS := filepath.Join(repoRoot, "script/debug/session-list-ux-enhance/probe.js")
	probeOut := filepath.Join(*outDir, "probe")
	_ = os.MkdirAll(probeOut, 0755)

	ctx := context.Background()
	report, raw, err := runProbe(ctx, probeJS, baseURL, probeOut, *token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: playwright probe: %v\nraw:\n%s\n", err, raw)
		os.Exit(2)
	}
	_ = os.WriteFile(filepath.Join(*outDir, "probe-report.json"), []byte(mustJSON(report)), 0644)

	fmt.Fprintf(os.Stderr, "probe: symptomPresent=%v okHealthy=%v issues=%v reasons=%v\n",
		report.SymptomPresent, report.OkHealthy, report.Issues, report.Reasons)
	for _, s := range report.Screenshots {
		fmt.Fprintf(os.Stderr, "screenshot: %s\n", s)
	}

	switch mode {
	case "repro":
		if report.SymptomPresent {
			fmt.Println("REPRO: session list UX defects present (load-more / auto-load / scroll preserve / snap-back)")
			for _, r := range report.Reasons {
				fmt.Printf("REPRO: %s\n", r)
			}
			if report.Symptom != nil {
				fmt.Printf("REPRO: symptom %s\n", compactJSON(report.Symptom))
			}
			for _, issue := range report.Issues {
				fmt.Printf("REPRO: issue: %s\n", issue)
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
		fmt.Fprintf(os.Stderr, "FAIL: expected UX symptoms not observed (cannot establish REPRO)\n")
		fmt.Fprintf(os.Stderr, "issues: %v\nokHealthy=%v\n", report.Issues, report.OkHealthy)
		fmt.Fprintf(os.Stderr, "report: %s\n", filepath.Join(*outDir, "probe-report.json"))
		if report.OkHealthy {
			fmt.Fprintln(os.Stderr, "hint: list UX already healthy; use --expect=healthy for VERIFY")
		}
		if stopWeb != nil {
			stopWeb()
		}
		os.Exit(2)

	case "healthy":
		if report.SymptomPresent || !report.OkHealthy {
			fmt.Println("FAIL: session list UX not healthy")
			for _, r := range report.Reasons {
				fmt.Printf("FAIL: %s\n", r)
			}
			for _, issue := range report.Issues {
				fmt.Printf("FAIL: %s\n", issue)
			}
			if report.Symptom != nil {
				fmt.Printf("FAIL: symptom %s\n", compactJSON(report.Symptom))
			}
			for _, s := range report.Screenshots {
				fmt.Printf("FAIL: screenshot %s\n", s)
			}
			if stopWeb != nil {
				stopWeb()
			}
			os.Exit(1)
		}
		fmt.Println("VERIFY: session list UX healthy (load-more, scroll, no snap-back, no wasteful runners/status poll)")
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

func seedManySessions(home, runner string, count int) error {
	if err := os.MkdirAll(filepath.Join(home, "sessions"), 0755); err != nil {
		return err
	}
	// Flat layout so store does not expect nested runner dirs.
	if err := os.WriteFile(filepath.Join(home, "sessions", ".layout"), []byte(`{"version":2}`+"\n"), 0644); err != nil {
		return err
	}
	base := time.Now().UTC().Add(-2 * time.Hour)
	for i := 1; i <= count; i++ {
		id := fmt.Sprintf("home-list-%03d", i)
		sessDir := filepath.Join(home, "sessions", id)
		if err := os.MkdirAll(sessDir, 0755); err != nil {
			return err
		}
		// Newer sessions get higher updated_at so newest-first list is stable.
		updated := base.Add(time.Duration(i) * time.Minute)
		created := updated.Add(-2 * time.Minute)
		meta := map[string]any{
			"runner":         runner,
			"session_id":     id,
			"initial_prompt": fmt.Sprintf("Home list seed session %s for pagination UX", id),
			"status":         "idle",
			"workspace":      filepath.Join(home, "workspace"),
			"created_at":     created.Format(time.RFC3339),
			"updated_at":     updated.Format(time.RFC3339),
		}
		metaBytes, err := json.Marshal(meta)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(sessDir, "meta.json"), metaBytes, 0644); err != nil {
			return err
		}
		ts := updated.UnixMilli()
		events := fmt.Sprintf(
			`{"type":"message","role":"user","text":"Home list seed session %s","timestamp":%d}`+"\n",
			id, ts,
		)
		if err := os.WriteFile(filepath.Join(sessDir, "events.jsonl"), []byte(events), 0644); err != nil {
			return err
		}
	}
	return nil
}

func waitHasMore(ctx context.Context, baseURL, bearer string, timeout time.Duration) error {
	client := &http.Client{Timeout: 3 * time.Second}
	url := strings.TrimRight(baseURL, "/") + "/api/agent-run/sessions?limit=30&offset=0"
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
			if resp.StatusCode == http.StatusOK {
				var data struct {
					Total   int  `json:"total"`
					HasMore bool `json:"has_more"`
				}
				if json.Unmarshal(body, &data) == nil && data.Total >= seedSessionCount && data.HasMore {
					return nil
				}
			}
		}
		time.Sleep(150 * time.Millisecond)
	}
	return fmt.Errorf("sessions has_more not ready: %s", url)
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
	// Nested cmd module: packages like ./agent-run live under cmd/go.mod.
	args := []string{"build", "-C", "cmd", "-o", out, pkg}
	if strings.HasPrefix(pkg, "./agent/") || strings.HasPrefix(pkg, "github.com/") {
		args = []string{"build", "-o", out, pkg}
	}
	cmd := exec.Command("go", args...)
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

func compactJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}
