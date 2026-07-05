package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	lessflags "github.com/xhd2015/less-flags"
	"github.com/xhd2015/agent-pro/pkgs/ttyrunner"
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

	var sessionID string
	var release func()
	if opts.customSessionID != "" {
		release, err = ReserveCustomSessionID(home, opts.customSessionID)
		if err != nil {
			return err
		}
		sessionID = opts.customSessionID
	} else {
		sessionID, release, err = ReserveRegistrySessionID(home)
		if err != nil {
			return err
		}
	}
	defer release()

	serveToken := ttywatch.ServeSubcommand(opts.commandArgs)
	argv := append([]string{serveToken, sessionID}, opts.commandArgs...)
	cmd := exec.Command(os.Args[0], argv...)
	cmd.Env = os.Environ()
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return err
	}

	entry, err := waitForRegistryEntry(home, sessionID, 15*time.Second)
	if err != nil {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
		return err
	}
	release()

	debugLogf("run start session=%s listen=%s command=%v home=%s argv0=%s headless=%v detach=%v",
		sessionID, entry.ListenAddr, opts.commandArgs, home, os.Args[0], opts.headless, opts.detach)

	if opts.detach {
		fmt.Printf("session-id: %s\n", sessionID)
		return nil
	}

	if opts.headless {
		fmt.Printf("session-id: %s\n", sessionID)
		return waitHeadless(cmd, entry, home, sessionID, opts.commandArgs)
	}

	// Do not wait on the serve child: on Ctrl-] detach the parent must exit
	// while the __serve__ process keeps the session alive.
	detached, err := attachWriter(entry.ListenAddr, sessionID, "screen")
	if err != nil {
		return err
	}
	if !detached {
		RemoveRegistryIfMatch(home, sessionID, entry.ListenAddr, entry.PID)
	}
	return nil
}

const headlessWaitingLine = "waiting for program to exit..."

func forwardHeadlessInterrupt(entry *RegistryEntry, listenAddr, sessionID string, command []string) error {
	if isBareSleepCommand(command) {
		// Bare sleep(1) ignores host-forwarded terminal interrupts on Darwin;
		// skip PTY forwarding so the grace window can run full length.
		debugLogf("headless sigint skip forward for bare sleep")
		return nil
	}

	if entry.PID > 0 {
		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			pgid, err := servePTYChildPGID(entry.PID)
			if err == nil && pgid > 0 {
				if err := syscall.Kill(-pgid, syscall.SIGINT); err == nil {
					debugLogf("headless sigint killpg pgid=%d", pgid)
					return nil
				}
			}
			time.Sleep(50 * time.Millisecond)
		}
	}

	// Fallback: attached-run parity via PTY inject (Ctrl-C byte).
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if err := prepareSessionInjectMode(listenAddr, sessionID); err == nil {
			return ttyrunner.InjectInput(listenAddr, sessionID, []byte{0x03})
		}
		time.Sleep(50 * time.Millisecond)
	}
	return ttyrunner.InjectInput(listenAddr, sessionID, []byte{0x03})
}

func isBareSleepCommand(command []string) bool {
	return len(command) >= 1 && filepath.Base(command[0]) == "sleep"
}

func servePTYChildPGID(servePID int) (int, error) {
	child, err := firstChildPID(servePID)
	if err != nil {
		return 0, err
	}
	return syscall.Getpgid(child)
}

func firstChildPID(ppid int) (int, error) {
	for _, pgrep := range pgrepCandidates() {
		out, err := exec.Command(pgrep, "-P", strconv.Itoa(ppid)).Output()
		if err != nil {
			continue
		}
		fields := strings.Fields(string(out))
		if len(fields) == 0 {
			return 0, fmt.Errorf("no child for pid %d", ppid)
		}
		return strconv.Atoi(fields[0])
	}
	return 0, fmt.Errorf("pgrep unavailable")
}

func pgrepCandidates() []string {
	var out []string
	seen := make(map[string]struct{})
	add := func(p string) {
		if p == "" {
			return
		}
		if _, ok := seen[p]; ok {
			return
		}
		if _, err := os.Stat(p); err != nil {
			return
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	if p, err := exec.LookPath("pgrep"); err == nil {
		add(p)
	}
	add("/usr/bin/pgrep")
	add("/bin/pgrep")
	return out
}

func waitHeadless(cmd *exec.Cmd, entry *RegistryEntry, home, sessionID string, command []string) error {
	sigintCh := make(chan os.Signal, 1)
	signal.Notify(sigintCh, syscall.SIGINT)
	defer signal.Stop(sigintCh)

	waitDone := make(chan error, 1)
	go func() {
		waitDone <- cmd.Wait()
	}()

	select {
	case waitErr := <-waitDone:
		return exitCodeFromWait(waitErr)
	case <-sigintCh:
		return handleHeadlessSIGINT(cmd, entry, home, sessionID, command, waitDone)
	}
}

func handleHeadlessSIGINT(cmd *exec.Cmd, entry *RegistryEntry, home, sessionID string, command []string, waitDone <-chan error) error {
	if err := forwardHeadlessInterrupt(entry, entry.ListenAddr, sessionID, command); err != nil {
		debugLogf("headless sigint forward interrupt: %v", err)
	}

	const (
		graceWindow = 10 * time.Second
		logAfter    = 1 * time.Second
	)
	start := time.Now()
	logged := false

	graceTimer := time.NewTimer(graceWindow)
	defer graceTimer.Stop()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case waitErr := <-waitDone:
			return exitCodeFromWait(waitErr)
		case <-graceTimer.C:
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
			_ = cmd.Wait()
			RemoveRegistryIfMatch(home, sessionID, entry.ListenAddr, entry.PID)
			return &cliExitError{code: 1}
		case <-ticker.C:
			if !logged && time.Since(start) >= logAfter {
				fmt.Fprintln(os.Stderr, headlessWaitingLine)
				logged = true
			}
		}
	}
}

func exitCodeFromWait(err error) error {
	if err == nil {
		return nil
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