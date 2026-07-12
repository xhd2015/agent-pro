// open-bind-smoke: reproduce / verify agent-run run --open grok session bind.
//
// Modes:
//
//	(default) repro  — expect symptom: runner unbound after --open with real grok
//	                    (exit non-zero + REPRO: lines when symptom present)
//	--expect=healthy — expect bound runner_session_id after --open (VERIFY gate)
//
// Usage:
//
//	go run ./script/debug/open-bind-smoke
//	go run ./script/debug/open-bind-smoke --expect=healthy
//	go run ./script/debug/open-bind-smoke --agent-run /path/to/agent-run
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	expectHealthy := flag.String("expect", "repro", "repro | healthy")
	agentRun := flag.String("agent-run", "", "path to agent-run binary (default: agent-run on PATH)")
	sessionID := flag.String("session-id", "smoke-open-bind-loop", "session id")
	prompt := flag.String("prompt", "one word of France capital", "open prompt")
	timeout := flag.Duration("timeout", 45*time.Second, "max wait for open command")
	flag.Parse()

	bin := strings.TrimSpace(*agentRun)
	if bin == "" {
		var err error
		bin, err = exec.LookPath("agent-run")
		if err != nil {
			fmt.Fprintf(os.Stderr, "FAIL: agent-run not on PATH: %v\n", err)
			os.Exit(2)
		}
	}
	if _, err := os.Stat(bin); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: agent-run binary missing: %v\n", err)
		os.Exit(2)
	}

	home, err := os.MkdirTemp("", "agent-run-open-bind-smoke-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: temp home: %v\n", err)
		os.Exit(2)
	}
	// Keep home for evidence on failure; remove only on healthy success.
	cleanupHome := true
	defer func() {
		if cleanupHome {
			_ = os.RemoveAll(home)
		}
	}()

	// Prefer a real project directory. Using /tmp (or empty temp dirs) makes real
	// grok show a "Run Grok Build in a project directory?" picker and never emit
	// the user prompt into updates.jsonl — discovery cannot bind.
	workdir, err := os.Getwd()
	if err != nil || strings.TrimSpace(workdir) == "" {
		workdir = filepath.Join(home, "workspace")
		if mkErr := os.MkdirAll(workdir, 0755); mkErr != nil {
			fmt.Fprintf(os.Stderr, "FAIL: workspace: %v\n", mkErr)
			os.Exit(2)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	args := []string{
		"run",
		"--session-id=" + *sessionID,
		"--agent-runner=grok-tty",
		"--open",
		"--dir", workdir,
		*prompt,
	}
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = workdir
	cmd.Env = append(os.Environ(),
		"AGENT_RUN_HOME="+home,
		// Non-interactive attach so smoke does not need a controlling TTY.
		"AGENT_RUN_OPEN_ATTACH_INSTANT=1",
		// Match user default: do NOT set GROK_HOME (soft path / short budget).
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	exitCode := 0
	if runErr != nil {
		if ee, ok := runErr.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else if ctx.Err() != nil {
			fmt.Fprintf(os.Stderr, "FAIL: open timed out: %v\n", ctx.Err())
			fmt.Fprintf(os.Stderr, "stderr:\n%s\n", stderr.String())
			os.Exit(2)
		} else {
			fmt.Fprintf(os.Stderr, "FAIL: run open: %v\n", runErr)
			os.Exit(2)
		}
	}

	metaPath := filepath.Join(home, "sessions", "grok-tty", *sessionID, "meta.json")
	bindPath := filepath.Join(home, "sessions", "grok-tty", *sessionID, "bind.json")
	metaRaw, _ := os.ReadFile(metaPath)
	bindRaw, _ := os.ReadFile(bindPath)

	statusOut, statusEC := runStatus(bin, home, *sessionID)

	var meta struct {
		RunnerSessionID string `json:"runner_session_id"`
		Status          string `json:"status"`
	}
	_ = json.Unmarshal(metaRaw, &meta)
	var bind struct {
		State string `json:"state"`
		Error string `json:"error"`
	}
	_ = json.Unmarshal(bindRaw, &bind)

	runnerID := strings.TrimSpace(meta.RunnerSessionID)
	bound := runnerID != ""
	statusUnbound := strings.Contains(statusOut, "unbound") ||
		strings.Contains(statusOut, "missing runner_session_id") ||
		strings.Contains(statusOut, "runner session not bound")
	bindFailed := strings.EqualFold(bind.State, "failed") ||
		strings.Contains(bind.Error, "deadline exceeded")

	fmt.Printf("smoke_home=%s\n", home)
	fmt.Printf("agent_run=%s\n", bin)
	fmt.Printf("open_exit=%d\n", exitCode)
	fmt.Printf("runner_session_id=%q\n", runnerID)
	fmt.Printf("bind_state=%q bind_error=%q\n", bind.State, bind.Error)
	fmt.Printf("status_exit=%d\n", statusEC)
	fmt.Printf("--- status ---\n%s", statusOut)
	if !strings.HasSuffix(statusOut, "\n") {
		fmt.Println()
	}
	fmt.Printf("--- open stderr ---\n%s", stderr.String())
	if !strings.HasSuffix(stderr.String(), "\n") {
		fmt.Println()
	}

	switch strings.ToLower(strings.TrimSpace(*expectHealthy)) {
	case "healthy":
		// VERIFY: open should bind real grok session under default GROK_HOME path.
		if !bound {
			cleanupHome = false
			fmt.Printf("VERIFY: FAIL missing runner_session_id after --open (kept smoke_home=%s)\n", home)
			os.Exit(1)
		}
		if statusUnbound {
			cleanupHome = false
			fmt.Println("VERIFY: FAIL status still reports unbound")
			os.Exit(1)
		}
		if !strings.Contains(stderr.String(), "grok session") {
			cleanupHome = false
			fmt.Println("VERIFY: FAIL stderr missing grok session line")
			os.Exit(1)
		}
		fmt.Println("VERIFY: PASS bound after --open")
		os.Exit(0)
	default:
		// REPRO: user symptom — unbound after --open (default soft path).
		// Symptom present when: not bound AND (status unbound OR bind failed soft).
		if !bound && (statusUnbound || bindFailed || exitCode == 0 && !strings.Contains(stderr.String(), "grok session")) {
			fmt.Println("REPRO: runner_session_id missing after agent-run run --open")
			if bindFailed {
				fmt.Printf("REPRO: bind.json state=%s error=%s\n", bind.State, bind.Error)
			}
			if statusUnbound {
				fmt.Println("REPRO: status reports unbound / missing runner_session_id")
			}
			if exitCode == 0 && !strings.Contains(stderr.String(), "grok session") {
				fmt.Println("REPRO: open exited 0 without printing grok session lines")
			}
			// Non-zero so bug-repro gate treats unfixed system as REPRO PASS (symptom present).
			os.Exit(1)
		}
		if bound {
			fmt.Println("REPRO: FAIL symptom absent — session already bound (system may be fixed)")
			os.Exit(0)
		}
		fmt.Println("REPRO: FAIL unexpected state (neither clear symptom nor bound)")
		os.Exit(2)
	}
}

func runStatus(bin, home, sessionID string) (string, int) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "status", sessionID)
	cmd.Env = append(os.Environ(), "AGENT_RUN_HOME="+home)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	ec := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			ec = ee.ExitCode()
		} else {
			ec = 2
		}
	}
	return out.String(), ec
}
