package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/xhd2015/less-flags"
	"github.com/xhd2015/agent-pro/pkgs/ttywatch"
)

type cliExitError struct {
	code int
}

func (e *cliExitError) Error() string {
	return fmt.Sprintf("exit status %d", e.code)
}

type runOptions struct {
	headless        bool
	detach          bool
	customSessionID string
	commandArgs     []string
}

func parseRunArgs(args []string) (runOptions, error) {
	var opts runOptions
	if duplicateSessionIDFlag(args) {
		return opts, fmt.Errorf("run: duplicate --session-id flag")
	}

	var sessionIDPtr *string
	remain, err := lessflags.String("--session-id", &sessionIDPtr).
		Bool("--headless", &opts.headless).
		Bool("--detach", &opts.detach).
		StopOnFirstArg().
		Parse(args)
	if err != nil {
		return opts, err
	}
	if sessionIDPtr != nil {
		opts.customSessionID = *sessionIDPtr
	}
	if opts.headless && opts.detach {
		return opts, fmt.Errorf("run: --headless and --detach cannot be used together")
	}
	opts.commandArgs = remain
	if len(opts.commandArgs) == 0 {
		return opts, fmt.Errorf("usage: tty-watch run [--headless] [--detach] [--session-id <id>] <command> [args...]")
	}
	return opts, nil
}

func duplicateSessionIDFlag(args []string) bool {
	count := 0
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			break
		}
		if !strings.HasPrefix(arg, "-") {
			break
		}
		if arg == "--session-id" || strings.HasPrefix(arg, "--session-id=") {
			count++
		}
	}
	return count > 1
}

func runRun(args []string) error {
	opts, err := parseRunArgs(args)
	if err != nil {
		return err
	}

	home, err := TTYWatchHome()
	if err != nil {
		return err
	}

	result, err := ttywatch.HeadlessRun(context.Background(), ttywatch.HeadlessRunOptions{
		Home:       home,
		SessionID:  opts.customSessionID,
		Command:    opts.commandArgs,
		BinaryPath: os.Args[0],
		// KeepAlive must stay false for default CLI run: when true, __serve__
		// waits forever on ctx after the PTY child exits and becomes an orphan.
		// Intentional keep-tty is opt-in via agenttty KeepTerminalAlive.
		KeepAlive: false,
	})
	if err != nil {
		return err
	}

	debugLogf("run start session=%s listen=%s command=%v home=%s argv0=%s headless=%v detach=%v",
		result.SessionID, result.Entry.ListenAddr, opts.commandArgs, home, os.Args[0], opts.headless, opts.detach)

	if opts.detach {
		fmt.Printf("session-id: %s\n", result.SessionID)
		return nil
	}

	if opts.headless {
		fmt.Printf("session-id: %s\n", result.SessionID)
		return waitHeadless(result, opts.commandArgs)
	}

	// Do not wait on the serve child: on Ctrl-] detach the parent must exit
	// while the __serve__ process keeps the session alive.
	detached, err := ttywatch.AttachWriter(result.Entry.ListenAddr, result.SessionID, "screen")
	if err != nil {
		return err
	}
	if !detached {
		RemoveRegistryIfMatch(home, result.SessionID, result.Entry.ListenAddr, result.Entry.PID)
	}
	return nil
}

func waitHeadless(result *ttywatch.HeadlessRunResult, command []string) error {
	err := ttywatch.WaitHeadless(context.Background(), result, command)
	return exitCodeFromWait(err)
}

func exitCodeFromWait(err error) error {
	if err == nil {
		return nil
	}
	var exitStatus *ttywatch.ExitStatus
	if errors.As(err, &exitStatus) {
		if exitStatus.Code == 0 {
			return nil
		}
		return &cliExitError{code: exitStatus.Code}
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		code := exitErr.ExitCode()
		if code == 0 {
			return nil
		}
		return &cliExitError{code: code}
	}
	return err
}