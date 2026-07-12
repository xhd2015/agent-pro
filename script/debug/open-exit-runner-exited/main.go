// open-exit-runner-exited reproduces / verifies status after /exit inside grok-tty.
//
// User symptom (unfixed): after grok /exit, keep-alive serve stays up, terminal
// TCP still reachable, but agent child is gone. status still reports:
//   runner.exited: false
//   resume.ready: no  (runner still active)
//
// Modes:
//
//	(default) repro   — exit 1 + REPRO: when symptom present
//	--expect=healthy  — exit 0 when after /exit: exited=true and resume ready
//	                    (or process/terminal dead with exited=true)
//
// Usage:
//
//	go run ./script/debug/open-exit-runner-exited --agent-run "$(which agent-run)"
//	go run ./script/debug/open-exit-runner-exited --agent-run "$BIN" --expect=healthy
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
	expect := flag.String("expect", "repro", "repro | healthy")
	agentRun := flag.String("agent-run", "", "agent-run binary (default: PATH)")
	sessionID := flag.String("session-id", "smoke-open-exit-exited", "session id")
	prompt := flag.String("prompt", "one word of France capital", "initial open prompt")
	timeout := flag.Duration("timeout", 90*time.Second, "overall timeout")
	flag.Parse()

	bin := strings.TrimSpace(*agentRun)
	if bin == "" {
		var err error
		bin, err = exec.LookPath("agent-run")
		if err != nil {
			fmt.Fprintf(os.Stderr, "FAIL: agent-run not found: %v\n", err)
			os.Exit(2)
		}
	}

	home, err := os.MkdirTemp("", "agent-run-open-exit-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: temp home: %v\n", err)
		os.Exit(2)
	}
	// Keep on failure for evidence.
	keepHome := false
	defer func() {
		if !keepHome {
			_ = os.RemoveAll(home)
		}
	}()

	workdir, err := os.Getwd()
	if err != nil || strings.TrimSpace(workdir) == "" {
		workdir = filepath.Join(home, "ws")
		_ = os.MkdirAll(workdir, 0755)
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// 1) Open keep-alive TTY (instant attach for non-interactive).
	openArgs := []string{
		"run",
		"--session-id=" + *sessionID,
		"--agent-runner=grok-tty",
		"--open",
		"--dir", workdir,
		*prompt,
	}
	openCmd := exec.CommandContext(ctx, bin, openArgs...)
	openCmd.Dir = workdir
	openCmd.Env = append(os.Environ(),
		"AGENT_RUN_HOME="+home,
		"AGENT_RUN_OPEN_ATTACH_INSTANT=1",
	)
	var openOut, openErr bytes.Buffer
	openCmd.Stdout = &openOut
	openCmd.Stderr = &openErr
	if err := openCmd.Run(); err != nil && ctx.Err() != nil {
		keepHome = true
		fmt.Fprintf(os.Stderr, "FAIL: open timed out: %v\nstderr:\n%s\n", err, openErr.String())
		os.Exit(2)
	}
	// Open may exit 0 after attach instant even if bind soft-fails; require bound for this loop.
	meta := readMeta(home, *sessionID)
	if strings.TrimSpace(meta.RunnerSessionID) == "" {
		// Wait briefly for bind.json ok (background bind).
		waitBound(home, *sessionID, 25*time.Second)
		meta = readMeta(home, *sessionID)
	}
	if strings.TrimSpace(meta.RunnerSessionID) == "" {
		keepHome = true
		fmt.Fprintf(os.Stderr, "FAIL: pre-exit bind failed (no runner_session_id); stderr:\n%s\n", openErr.String())
		fmt.Fprintf(os.Stderr, "bind=%s\n", readFile(filepath.Join(home, "sessions", "grok-tty", *sessionID, "bind.json")))
		os.Exit(2)
	}

	// 2) Send /exit into live terminal (same as user exiting inside grok).
	// Wait until sendable when possible.
	waitSendable(bin, home, *sessionID, 30*time.Second)
	sendCmd := exec.CommandContext(ctx, bin, "send", *sessionID, "/exit")
	sendCmd.Env = append(os.Environ(), "AGENT_RUN_HOME="+home)
	var sendOut, sendErr bytes.Buffer
	sendCmd.Stdout = &sendOut
	sendCmd.Stderr = &sendErr
	_ = sendCmd.Run() // may fail if already exited; continue to status

	// Give serve a moment to leave zombie / scrollback footer.
	time.Sleep(2 * time.Second)

	// 3) Status probe (human + json).
	statusHuman, _ := runAgent(bin, home, 15*time.Second, "status", *sessionID)
	statusJSON, _ := runAgent(bin, home, 15*time.Second, "status", "--json", *sessionID)

	var st struct {
		Process struct {
			Status string `json:"status"`
			PID    int    `json:"pid"`
		} `json:"process"`
		Terminal struct {
			Status   string `json:"status"`
			Sendable string `json:"sendable"`
		} `json:"terminal"`
		Runner struct {
			Status    string `json:"status"`
			SessionID string `json:"session_id"`
			Exited    *bool  `json:"exited"`
		} `json:"runner"`
		Resume struct {
			Ready  bool   `json:"ready"`
			Reason string `json:"reason"`
		} `json:"resume"`
		Status string `json:"status"`
	}
	_ = json.Unmarshal([]byte(statusJSON), &st)

	exitedFalse := st.Runner.Exited != nil && !*st.Runner.Exited
	exitedTrue := st.Runner.Exited != nil && *st.Runner.Exited
	termReachable := st.Terminal.Status == "reachable" || strings.Contains(statusHuman, "reachable")
	procAlive := st.Process.Status == "alive" || strings.Contains(statusHuman, "process:\n  status:  alive")
	resumeBlocked := !st.Resume.Ready && (strings.Contains(st.Resume.Reason, "still active") ||
		strings.Contains(statusHuman, "runner still active"))

	// Child gone evidence: serve may still be alive; try listing children of process pid.
	childGone := true
	if st.Process.PID > 0 {
		if out, err := exec.Command("pgrep", "-P", fmt.Sprint(st.Process.PID)).CombinedOutput(); err == nil && len(bytes.TrimSpace(out)) > 0 {
			childGone = false
		}
	}

	fmt.Printf("smoke_home=%s\n", home)
	fmt.Printf("agent_run=%s\n", bin)
	fmt.Printf("runner_session_id=%q\n", meta.RunnerSessionID)
	fmt.Printf("send_stdout=%q send_stderr=%q\n", strings.TrimSpace(sendOut.String()), strings.TrimSpace(sendErr.String()))
	fmt.Printf("process_status=%q terminal_status=%q sendable=%q\n", st.Process.Status, st.Terminal.Status, st.Terminal.Sendable)
	fmt.Printf("runner_exited=%v resume_ready=%v resume_reason=%q child_gone=%v\n",
		ptrBool(st.Runner.Exited), st.Resume.Ready, st.Resume.Reason, childGone)
	fmt.Printf("--- status ---\n%s", statusHuman)
	if !strings.HasSuffix(statusHuman, "\n") {
		fmt.Println()
	}
	fmt.Printf("--- open stderr ---\n%s", openErr.String())
	if !strings.HasSuffix(openErr.String(), "\n") {
		fmt.Println()
	}

	// Symptom: after /exit, serve may stay alive and TCP reachable, but agent
	// is gone — yet status claims exited:false and blocks resume.
	symptom := exitedFalse && termReachable && resumeBlocked
	// Stronger: also process alive + not sendable (zombie serve after child exit).
	if procAlive && strings.Contains(strings.ToLower(st.Terminal.Sendable), "no") {
		symptom = symptom || (exitedFalse && termReachable)
	}

	switch strings.ToLower(strings.TrimSpace(*expect)) {
	case "healthy":
		// After /exit: runner.exited true and resume ready (bound+exited), OR
		// terminal unreachable with exited true.
		ok := exitedTrue && (st.Resume.Ready || st.Runner.Status == "bound")
		if !ok {
			// Also accept: process dead + exited true + resume ready
			ok = exitedTrue && st.Process.Status == "dead" && st.Resume.Ready
		}
		if !ok {
			keepHome = true
			fmt.Println("VERIFY: FAIL after /exit expected runner.exited=true and resume.ready=true")
			os.Exit(1)
		}
		fmt.Println("VERIFY: PASS after /exit runner.exited=true and resume allows resume")
		os.Exit(0)
	default:
		if symptom {
			keepHome = true
			fmt.Println("REPRO: after /exit status reports runner.exited=false while terminal still reachable")
			fmt.Println("REPRO: resume.ready=no reason claims runner still active")
			if procAlive {
				fmt.Println("REPRO: process.status=alive (keep-alive serve zombie)")
			}
			if childGone {
				fmt.Println("REPRO: serve has no child process (grok already exited)")
			}
			os.Exit(1)
		}
		if exitedTrue && st.Resume.Ready {
			fmt.Println("REPRO: FAIL symptom absent — exited=true and resume ready (system may be fixed)")
			os.Exit(0)
		}
		keepHome = true
		fmt.Println("REPRO: FAIL unexpected state (could not establish open+/exit scenario cleanly)")
		os.Exit(2)
	}
}

type metaFile struct {
	RunnerSessionID string `json:"runner_session_id"`
	Status          string `json:"status"`
}

func readMeta(home, sessionID string) metaFile {
	var m metaFile
	b, err := os.ReadFile(filepath.Join(home, "sessions", "grok-tty", sessionID, "meta.json"))
	if err != nil {
		return m
	}
	_ = json.Unmarshal(b, &m)
	return m
}

func readFile(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(b)
}

func waitBound(home, sessionID string, max time.Duration) {
	deadline := time.Now().Add(max)
	for time.Now().Before(deadline) {
		m := readMeta(home, sessionID)
		if strings.TrimSpace(m.RunnerSessionID) != "" {
			return
		}
		// bind.json ok
		var st struct {
			State string `json:"state"`
		}
		_ = json.Unmarshal([]byte(readFile(filepath.Join(home, "sessions", "grok-tty", sessionID, "bind.json"))), &st)
		if st.State == "ok" {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func waitSendable(bin, home, sessionID string, max time.Duration) {
	deadline := time.Now().Add(max)
	for time.Now().Before(deadline) {
		out, _ := runAgent(bin, home, 10*time.Second, "status", "--json", sessionID)
		if strings.Contains(out, `"sendable":"yes"`) || strings.Contains(out, `"sendable": "yes"`) {
			return
		}
		// also human
		hout, _ := runAgent(bin, home, 10*time.Second, "status", sessionID)
		if strings.Contains(hout, "sendable: yes") {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func runAgent(bin, home string, timeout time.Duration, args ...string) (string, int) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
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
			ec = -1
		}
	}
	return out.String(), ec
}

func ptrBool(p *bool) string {
	if p == nil {
		return "null"
	}
	if *p {
		return "true"
	}
	return "false"
}
