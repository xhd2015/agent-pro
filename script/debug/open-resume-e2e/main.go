// open-resume-e2e: end-to-end health check for open → Paris → /exit → resume --open.
//
// Flow:
//  1. agent-run run --session-id=… --agent-runner=grok-tty --open "one word of France capital"
//  2. wait until snapshot contains "Paris"
//  3. send "/exit"
//  4. wait until status runner.exited=true
//  5. agent-run resume <id> --open "hello"
//  6. snapshot contains proper text (hello and/or resumed UI, not empty/error-only)
//
// Usage:
//
//	go run ./script/debug/open-resume-e2e
//	go run ./script/debug/open-resume-e2e --agent-run /path/to/agent-run
//	go run ./script/debug/open-resume-e2e --session-id=test-open-v8
//
// Exit 0 = VERIFY PASS (healthy). Non-zero = FAIL with reason.
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
	agentRun := flag.String("agent-run", "", "agent-run binary (default: PATH)")
	sessionID := flag.String("session-id", "test-open-v8", "session id")
	prompt := flag.String("prompt", "one word of France capital", "initial open prompt")
	followup := flag.String("followup", "hello", "resume followup")
	wantParis := flag.String("want-paris", "Paris", "substring expected in first-turn snapshot")
	timeout := flag.Duration("timeout", 3*time.Minute, "overall timeout")
	parisWait := flag.Duration("paris-wait", 90*time.Second, "max wait for Paris in snapshot")
	exitWait := flag.Duration("exit-wait", 45*time.Second, "max wait for exited:true after /exit")
	flag.Parse()

	bin := strings.TrimSpace(*agentRun)
	if bin == "" {
		var err error
		bin, err = exec.LookPath("agent-run")
		if err != nil {
			fail(2, "agent-run not on PATH: %v", err)
		}
	}
	if _, err := os.Stat(bin); err != nil {
		fail(2, "agent-run binary missing: %v", err)
	}
	if _, err := exec.LookPath("grok"); err != nil {
		fail(2, "grok not on PATH: %v", err)
	}

	home, err := os.MkdirTemp("", "agent-run-open-resume-e2e-*")
	if err != nil {
		fail(2, "temp home: %v", err)
	}
	keepHome := false
	defer func() {
		if !keepHome {
			_ = os.RemoveAll(home)
		} else {
			fmt.Printf("EVIDENCE: smoke_home=%s (kept on failure)\n", home)
		}
	}()

	// Project workspace (not bare /tmp — avoids grok project-directory picker).
	workdir, err := os.Getwd()
	if err != nil || strings.TrimSpace(workdir) == "" {
		workdir = filepath.Join(home, "ws")
		_ = os.MkdirAll(workdir, 0755)
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	env := append(os.Environ(),
		"AGENT_RUN_HOME="+home,
		// Non-interactive attach for automation.
		"AGENT_RUN_OPEN_ATTACH_INSTANT=1",
	)

	fmt.Printf("CHECK: open session %s\n", *sessionID)
	fmt.Printf("agent_run=%s\n", bin)
	fmt.Printf("AGENT_RUN_HOME=%s\n", home)
	fmt.Printf("workspace=%s\n", workdir)

	// 1) Open first turn (keep-alive). Instant attach returns after inject.
	openArgs := []string{
		"run",
		"--session-id=" + *sessionID,
		"--agent-runner=grok-tty",
		"--open",
		"--dir", workdir,
		*prompt,
	}
	_, openErr, openEC := runCmd(ctx, bin, workdir, env, openArgs...)
	fmt.Printf("open_exit=%d\n", openEC)
	if openErr != "" {
		fmt.Printf("--- open stderr ---\n%s\n", openErr)
	}
	if openEC != 0 && !strings.Contains(openErr, "grok session") {
		// Bind may still succeed with exit 0; hard-fail if nothing useful.
		keepHome = true
		fail(1, "FAIL: open exited %d without successful session bind; stderr=%s", openEC, openErr)
	}

	// 2) Wait for Paris in snapshot.
	fmt.Printf("CHECK: wait snapshot contains %q\n", *wantParis)
	parisSnap, err := waitSnapshotContains(ctx, bin, env, *sessionID, *wantParis, *parisWait)
	if err != nil {
		keepHome = true
		fail(1, "FAIL: Paris not in snapshot: %v\nlast_snapshot:\n%s", err, parisSnap)
	}
	fmt.Printf("PASS: snapshot contains %q\n", *wantParis)

	// 3) Send /exit.
	fmt.Printf("CHECK: send /exit\n")
	// Prefer sendable when possible, but /exit may still enqueue.
	waitSendable(ctx, bin, env, *sessionID, 30*time.Second)
	sendOut, sendErr, sendEC := runCmd(ctx, bin, workdir, env, "send", *sessionID, "/exit")
	fmt.Printf("send_exit=%d stdout=%q stderr=%q\n", sendEC, strings.TrimSpace(sendOut), strings.TrimSpace(sendErr))
	// Continue even if send fails — status wait is the gate.

	// 4) Wait exited:true.
	fmt.Printf("CHECK: wait status runner.exited=true\n")
	statusHuman, err := waitExitedTrue(ctx, bin, env, *sessionID, *exitWait)
	if err != nil {
		keepHome = true
		fail(1, "FAIL: exited:true not reached: %v\nlast_status:\n%s", err, statusHuman)
	}
	fmt.Printf("PASS: runner.exited=true\n")
	fmt.Printf("--- status after exit ---\n%s\n", statusHuman)

	// 5) Resume --open "hello".
	fmt.Printf("CHECK: resume --open %q\n", *followup)
	_, resErr, resEC := runCmd(ctx, bin, workdir, env,
		"resume", *sessionID, "--open", *followup)
	fmt.Printf("resume_exit=%d\n", resEC)
	if resErr != "" {
		fmt.Printf("--- resume stderr ---\n%s\n", resErr)
	}
	if resEC != 0 {
		// "already in use" is a known failure mode.
		keepHome = true
		fail(1, "FAIL: resume exited %d; stderr=%s", resEC, resErr)
	}
	if strings.Contains(strings.ToLower(resErr), "already in use") {
		keepHome = true
		fail(1, "FAIL: resume reported already in use")
	}

	// 6) Snapshot after resume must show proper text (not empty / not only exit footer).
	fmt.Printf("CHECK: snapshot after resume has proper text\n")
	// Give TUI a moment to render followup / resume UI.
	time.Sleep(2 * time.Second)
	snap, _, snapEC := runCmd(ctx, bin, workdir, env, "snapshot", *sessionID)
	fmt.Printf("snapshot_exit=%d\n", snapEC)
	fmt.Printf("--- resume snapshot ---\n%s\n", snap)

	if strings.TrimSpace(snap) == "" {
		keepHome = true
		fail(1, "FAIL: resume snapshot empty")
	}
	// Proper text: followup visible and/or interactive prompt; not only Terminal exited.
	lower := strings.ToLower(snap)
	onlyExited := strings.Contains(lower, "terminal exited") &&
		!strings.Contains(snap, *followup) &&
		!strings.Contains(lower, "hello") &&
		!strings.Contains(snap, "❯") &&
		!strings.Contains(lower, "grok")
	if onlyExited {
		keepHome = true
		fail(1, "FAIL: snapshot looks like dead/exited terminal only")
	}
	// Prefer seeing the followup "hello" after resume inject+submit.
	// Allow soft pass if resumed UI is interactive even before hello appears.
	hasHello := strings.Contains(snap, *followup) || strings.Contains(lower, "hello")
	hasUI := strings.Contains(snap, "❯") || strings.Contains(lower, "enter:send") ||
		strings.Contains(lower, "resume") || strings.Contains(snap, "Grok")
	if !hasHello && !hasUI {
		keepHome = true
		fail(1, "FAIL: snapshot missing followup %q and interactive UI markers", *followup)
	}
	if hasHello {
		fmt.Printf("PASS: snapshot contains followup %q\n", *followup)
	} else {
		fmt.Printf("PASS: snapshot shows interactive resumed UI (followup may still be rendering)\n")
	}

	// Cleanup serve processes for this session home.
	cleanupRegistryServes(home)

	fmt.Println("RESULT: PASS")
	fmt.Println("VERIFY: PASS open→Paris→/exit→exited→resume --open→snapshot ok")
	os.Exit(0)
}

func fail(code int, format string, args ...any) {
	fmt.Printf("RESULT: FAIL\n")
	fmt.Printf(format+"\n", args...)
	os.Exit(code)
}

func runCmd(ctx context.Context, bin, dir string, env []string, args ...string) (stdout, stderr string, exitCode int) {
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = dir
	cmd.Env = env
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else if ctx.Err() != nil {
			exitCode = -1
		} else {
			exitCode = 1
		}
	}
	return out.String(), errb.String(), exitCode
}

func waitSnapshotContains(ctx context.Context, bin string, env []string, sessionID, want string, max time.Duration) (string, error) {
	deadline := time.Now().Add(max)
	var last string
	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			return last, ctx.Err()
		}
		out, _, _ := runCmd(ctx, bin, "", env, "snapshot", sessionID)
		last = out
		if strings.Contains(out, want) {
			return out, nil
		}
		// Case-insensitive fallback for "paris"
		if strings.Contains(strings.ToLower(out), strings.ToLower(want)) {
			return out, nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return last, fmt.Errorf("timeout after %s waiting for %q", max, want)
}

func waitSendable(ctx context.Context, bin string, env []string, sessionID string, max time.Duration) {
	deadline := time.Now().Add(max)
	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			return
		}
		out, _, _ := runCmd(ctx, bin, "", env, "status", sessionID)
		if strings.Contains(out, "sendable: yes") {
			return
		}
		time.Sleep(400 * time.Millisecond)
	}
}

func waitExitedTrue(ctx context.Context, bin string, env []string, sessionID string, max time.Duration) (string, error) {
	deadline := time.Now().Add(max)
	var last string
	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			return last, ctx.Err()
		}
		human, _, _ := runCmd(ctx, bin, "", env, "status", sessionID)
		last = human
		if strings.Contains(human, "exited:     true") || strings.Contains(human, "exited: true") {
			return human, nil
		}
		// JSON path
		js, _, _ := runCmd(ctx, bin, "", env, "status", "--json", sessionID)
		var st struct {
			Runner struct {
				Exited *bool `json:"exited"`
			} `json:"runner"`
		}
		if json.Unmarshal([]byte(js), &st) == nil && st.Runner.Exited != nil && *st.Runner.Exited {
			return human, nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return last, fmt.Errorf("timeout after %s", max)
}

func cleanupRegistryServes(home string) {
	for _, sub := range []string{"grok-tty-registry", "codex-tty-registry"} {
		dir := filepath.Join(home, sub)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			b, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				continue
			}
			var ent struct {
				PID int `json:"pid"`
			}
			if json.Unmarshal(b, &ent) != nil || ent.PID <= 0 {
				continue
			}
			if ent.PID == os.Getpid() {
				continue
			}
			_ = exec.Command("kill", fmt.Sprint(ent.PID)).Run()
		}
	}
}
