package agentrunbridge

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	defaultBinary            = "agent-run"
	defaultReadyTimeout      = 60 * time.Second
	defaultReadyPollInterval = 500 * time.Millisecond
)

// RunOpts is the full-config entrypoint for launching agent-run.
type RunOpts struct {
	Prompt, SessionID, Binary, AgentRunner, RunnerConfigHome, WorkspaceDir string
	AutoSendOrResume, KeepTTY, NewTerminal, Open, NoSubmit, Stateless      bool
	WaitReady                                                              bool
	ReadyTimeout, ReadyPollInterval                                        time.Duration
	CaptureStdout                                                          bool
	LookPath                                                               func(file string) (string, error)
	RunCommand                                                             func(name string, args ...string) error
	RunOutput                                                              func(name string, args ...string) (string, error)
	Logf                                                                   func(format string, args ...any)
}

// RunResult holds optional captured launch stdout.
type RunResult struct {
	Stdout string
}

// Run resolves the binary, builds argv, launches agent-run, and optionally
// polls tty status until the session is ready.
func Run(opts RunOpts) (RunResult, error) {
	var result RunResult

	prompt := strings.TrimSpace(opts.Prompt)
	if prompt == "" {
		return result, fmt.Errorf("empty prompt")
	}
	// Use trimmed prompt for argv.
	opts.Prompt = prompt

	binary := strings.TrimSpace(opts.Binary)
	if binary == "" {
		binary = defaultBinary
	}

	lookPath := opts.LookPath
	if lookPath == nil {
		lookPath = exec.LookPath
	}
	resolved, err := lookPath(binary)
	if err != nil {
		return result, fmt.Errorf("%s not found on PATH: %w", binary, err)
	}
	if resolved == "" {
		resolved = binary
	}

	args := BuildArgs(opts)

	runCommand := opts.RunCommand
	if runCommand == nil {
		runCommand = defaultRunCommand
	}
	runOutput := opts.RunOutput
	if runOutput == nil {
		runOutput = defaultRunOutput
	}

	if opts.CaptureStdout {
		stdout, err := runOutput(resolved, args...)
		if err != nil {
			return result, err
		}
		result.Stdout = strings.TrimSpace(stdout)
	} else {
		if err := runCommand(resolved, args...); err != nil {
			return result, err
		}
	}

	if opts.WaitReady {
		if err := waitSessionReady(opts, resolved, runOutput); err != nil {
			return result, err
		}
	}
	return result, nil
}

func waitSessionReady(opts RunOpts, binary string, runOutput func(name string, args ...string) (string, error)) error {
	timeout := opts.ReadyTimeout
	if timeout <= 0 {
		timeout = defaultReadyTimeout
	}
	interval := opts.ReadyPollInterval
	if interval <= 0 {
		interval = defaultReadyPollInterval
	}

	sessionID := strings.TrimSpace(opts.SessionID)
	deadline := time.Now().Add(timeout)
	logf := opts.Logf
	if logf == nil {
		logf = func(string, ...any) {}
	}

	var lastStdout string
	var lastErr error
	for {
		stdout, err := runOutput(binary, "tty", "status", sessionID)
		lastStdout, lastErr = stdout, err
		if err == nil && IsSessionReady(stdout) {
			return nil
		}
		if !time.Now().Before(deadline) {
			screen, sendable := ParseTTYStatus(lastStdout)
			logf("ready timeout session=%s screen=%s sendable=%s err=%v", sessionID, screen, sendable, lastErr)
			return fmt.Errorf(
				"ready timeout after %s: session %q not banner+sendable (last status err=%v stdout=%q)",
				timeout, sessionID, lastErr, truncateForErr(lastStdout, 200),
			)
		}
		sleep := interval
		if rem := time.Until(deadline); rem < sleep {
			if rem <= 0 {
				screen, sendable := ParseTTYStatus(lastStdout)
				logf("ready timeout session=%s screen=%s sendable=%s err=%v", sessionID, screen, sendable, lastErr)
				return fmt.Errorf(
					"ready timeout after %s: session %q not banner+sendable (last status err=%v stdout=%q)",
					timeout, sessionID, lastErr, truncateForErr(lastStdout, 200),
				)
			}
			sleep = rem
		}
		time.Sleep(sleep)
	}
}

func defaultRunCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	// Inherit streams for interactive --open / operator visibility (SeaTalk parity).
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func defaultRunOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func truncateForErr(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
