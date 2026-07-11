package ttywatch

import (
	"context"
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
)

const (
	headlessRegistryTimeout = 15 * time.Second
	// HeadlessWaitingLine is logged to stderr once during the SIGINT grace window.
	HeadlessWaitingLine = "waiting for program to exit..."

	envTTYWatchRegistrySubdir = "TTY_WATCH_REGISTRY_SUBDIR"
	envTTYWatchExtraPaths     = "TTY_WATCH_EXTRA_PATHS"
	envTTYWatchKeepAlive      = "TTY_WATCH_KEEP_ALIVE"
)

// HeadlessRunOptions configures a detached __serve__ child session.
type HeadlessRunOptions struct {
	Home           string
	RegistrySubdir string // e.g. "grok-tty-registry"; empty uses tty-watch default
	SessionID      string // empty → auto-reserve session-N
	Command        []string
	BinaryPath     string // re-exec target for __serve__ (os.Args[0])
	Cwd            string
	ExtraPaths     []string
	// KeepAlive, when true, sets TTY_WATCH_KEEP_ALIVE so __serve__ stays alive
	// after the PTY child exits (intentional keep-tty). Default false: serve
	// exits shortly after the child so short attached runs leave no orphans.
	KeepAlive bool
}

// HeadlessRunResult is returned after the detached serve child registers.
type HeadlessRunResult struct {
	SessionID string
	Entry     *RegistryEntry
	Cmd       *exec.Cmd
	Registry  RegistryConfig
	Wait      func() error
}

// ExitStatus reports a non-zero exit code from headless wait handling.
type ExitStatus struct {
	Code int
}

func (e *ExitStatus) Error() string {
	return fmt.Sprintf("exit status %d", e.Code)
}

// HeadlessRun reserves a session id, spawns a detached __serve__ child, and waits
// for the registry entry.
func HeadlessRun(ctx context.Context, opts HeadlessRunOptions) (*HeadlessRunResult, error) {
	if len(opts.Command) == 0 {
		return nil, fmt.Errorf("headless run: missing command")
	}
	binaryPath := opts.BinaryPath
	if binaryPath == "" {
		return nil, fmt.Errorf("headless run: missing binary path")
	}

	home := opts.Home
	if home == "" {
		var err error
		home, err = TTYWatchHome()
		if err != nil {
			return nil, err
		}
	}
	cfg := registryConfigFor(home, opts.RegistrySubdir)

	var sessionID string
	var release func()
	var err error
	if opts.SessionID != "" {
		release, err = ReserveCustomSessionID(cfg, opts.SessionID)
		if err != nil {
			return nil, err
		}
		sessionID = opts.SessionID
	} else {
		sessionID, release, err = ReserveRegistrySessionID(cfg)
		if err != nil {
			return nil, err
		}
	}
	// Reservation wrote a provisional claim under the flock. Release immediately
	// so other runs are not blocked for the whole WaitForRegistryEntry window.
	release()

	serveToken := ServeSubcommand(opts.Command)
	argv := append([]string{serveToken, sessionID}, opts.Command...)
	cmd := exec.CommandContext(ctx, binaryPath, argv...)
	cmd.Env = serveChildEnv(os.Environ(), opts)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	if opts.Cwd != "" {
		cmd.Dir = opts.Cwd
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		clearSessionClaim(cfg, sessionID)
		return nil, err
	}

	entry, err := WaitForRegistryEntry(cfg, sessionID, headlessRegistryTimeout)
	if err != nil {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
		// Drop claim and any partial registry left by a failed serve.
		RemoveRegistry(cfg, sessionID)
		return nil, err
	}

	return &HeadlessRunResult{
		SessionID: sessionID,
		Entry:     entry,
		Cmd:       cmd,
		Registry:  cfg,
		Wait: func() error {
			return cmd.Wait()
		},
	}, nil
}

// WaitHeadless blocks until the serve child exits, forwarding SIGINT to the PTY
// session with the same grace window as tty-watch run --headless.
func WaitHeadless(ctx context.Context, result *HeadlessRunResult, command []string) error {
	if result == nil || result.Cmd == nil {
		return fmt.Errorf("headless wait: missing serve child")
	}
	cmd := result.Cmd

	sigintCh := make(chan os.Signal, 1)
	signal.Notify(sigintCh, syscall.SIGINT)
	defer signal.Stop(sigintCh)

	waitDone := make(chan error, 1)
	go func() {
		waitDone <- result.Wait()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case waitErr := <-waitDone:
		return exitStatusFromWait(waitErr)
	case <-sigintCh:
		return handleHeadlessSIGINT(ctx, cmd, result.Entry, result.Registry, result.SessionID, command, waitDone)
	}
}

func registryConfigFor(home, subdir string) RegistryConfig {
	if subdir == "" {
		return DefaultRegistryConfig(home)
	}
	return RegistryConfig{Home: home, Subdir: subdir}
}

func serveChildEnv(base []string, opts HeadlessRunOptions) []string {
	env := append([]string(nil), base...)
	if opts.Home != "" {
		env = setEnvVar(env, envTTYWatchHome, opts.Home)
	}
	if opts.RegistrySubdir != "" {
		env = setEnvVar(env, envTTYWatchRegistrySubdir, opts.RegistrySubdir)
	}
	if len(opts.ExtraPaths) > 0 {
		env = setEnvVar(env, envTTYWatchExtraPaths, strings.Join(opts.ExtraPaths, string(os.PathListSeparator)))
	}
	if opts.KeepAlive {
		env = setEnvVar(env, envTTYWatchKeepAlive, "1")
	} else {
		// Explicitly clear ambient TTY_WATCH_KEEP_ALIVE so a parent process that
		// exported keep-alive does not force detached serve children to linger.
		env = withoutEnvVar(env, envTTYWatchKeepAlive)
	}
	return env
}

func withoutEnvVar(env []string, key string) []string {
	prefix := key + "="
	out := make([]string, 0, len(env))
	for _, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			continue
		}
		out = append(out, entry)
	}
	return out
}

func setEnvVar(env []string, key, value string) []string {
	prefix := key + "="
	out := make([]string, 0, len(env)+1)
	replaced := false
	for _, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			out = append(out, prefix+value)
			replaced = true
			continue
		}
		out = append(out, entry)
	}
	if !replaced {
		out = append(out, prefix+value)
	}
	return out
}

func serveRegistryConfig(home, registrySubdir string) RegistryConfig {
	subdir := registrySubdir
	if subdir == "" {
		subdir = os.Getenv(envTTYWatchRegistrySubdir)
	}
	return registryConfigFor(home, subdir)
}

func serveExtraPaths(opts []string) []string {
	if len(opts) > 0 {
		return append([]string(nil), opts...)
	}
	raw := os.Getenv(envTTYWatchExtraPaths)
	if raw == "" {
		return nil
	}
	return strings.Split(raw, string(os.PathListSeparator))
}

func exitStatusFromWait(err error) error {
	if err == nil {
		return nil
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return err
	}
	code := exitErr.ExitCode()
	if code == 0 {
		return nil
	}
	return &ExitStatus{Code: code}
}

func handleHeadlessSIGINT(ctx context.Context, cmd *exec.Cmd, entry *RegistryEntry, cfg RegistryConfig, sessionID string, command []string, waitDone <-chan error) error {
	if entry == nil {
		return fmt.Errorf("headless sigint: missing registry entry")
	}
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
		case <-ctx.Done():
			return ctx.Err()
		case waitErr := <-waitDone:
			return exitStatusFromWait(waitErr)
		case <-graceTimer.C:
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
			_ = cmd.Wait()
			RemoveRegistryIfMatch(cfg, sessionID, entry.ListenAddr, entry.PID)
			return &ExitStatus{Code: 1}
		case <-ticker.C:
			if !logged && time.Since(start) >= logAfter {
				fmt.Fprintln(os.Stderr, HeadlessWaitingLine)
				logged = true
			}
		}
	}
}

func forwardHeadlessInterrupt(entry *RegistryEntry, listenAddr, sessionID string, command []string) error {
	if isBareSleepCommand(command) {
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

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if err := PrepareSessionInjectMode(listenAddr, sessionID); err == nil {
			return InjectInput(listenAddr, sessionID, []byte{0x03})
		}
		time.Sleep(50 * time.Millisecond)
	}
	return InjectInput(listenAddr, sessionID, []byte{0x03})
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