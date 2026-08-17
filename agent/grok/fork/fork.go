package fork

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	groksessions "github.com/xhd2015/agent-pro/agent/grok/sessions"
	"github.com/xhd2015/agent-pro/pkgs/procresolve"
	"github.com/xhd2015/agent-pro/pkgs/shell"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
	lessflags "github.com/xhd2015/less-flags"
	"golang.org/x/term"
)

const usage = `
Usage: grok-fork [OPTIONS]

Fork a Grok CLI session.

Without --session-id, resolve the nearest ancestor grok and open a new
iTerm2 window that re-invokes grok-fork --session-id <id>.

With --session-id, run grok --resume <id> --fork-session in the current
terminal.

Options:
  --session-id ID   fork the given grok session in the current terminal
  --dir DIR         workspace (default: session info.cwd)
  --pid PID         start pid for ancestor walk
  --dry-run         print plan only; do not launch
  --color           force ANSI color on
  --no-color        force ANSI color off
  -h,--help         show help
`

const (
	ansiReset = "\x1b[0m"
	ansiGreen = "\x1b[32m"
	ansiGray  = "\x1b[90m"
)

// Options configures Main. Hooks and homes are injectable; Main does not
// read GROK_HOME from the process environment.
type Options struct {
	Stdout, Stderr    io.Writer
	PID               int
	GrokHome          string
	ListProcs         func() []procresolve.Proc
	Lsof              func(pid int) []string
	OpenInNewTerminal func(dir, followUp string) error
	RunForeground     func(bin string, argv []string, dir string, env []string) error
	GrokBin           string
	LookPath          func(file string) (string, error)
	Executable        func() (string, error)
	Env               []string
}

// Run is an alias for Main.
func Run(args []string, opts *Options) error {
	return Main(args, opts)
}

// Main parses argv (no binary name) and runs Mode A (new window) or Mode B
// (current-terminal grok --resume --fork-session). It does not os.Exit and
// does not print an Error: prefix.
func Main(args []string, opts *Options) error {
	if opts == nil {
		opts = &Options{}
	}
	applyDefaults(opts)

	var sessionIDFlag *string
	var dirFlag *string
	var pidFlag *int
	var dryRun bool
	var color bool
	var noColor bool

	stdout := opts.Stdout
	remaining, err := lessflags.String("--session-id", &sessionIDFlag).
		String("--dir", &dirFlag).
		Int("--pid", &pidFlag).
		Bool("--dry-run", &dryRun).
		Bool("--color", &color).
		Bool("--no-color", &noColor).
		HelpFunc("-h,--help", func() {
			txt := strings.TrimPrefix(usage, "\n")
			fmt.Fprint(stdout, txt)
			if !strings.HasSuffix(txt, "\n") {
				fmt.Fprintln(stdout)
			}
		}).
		HelpNoExit().
		Parse(args)
	if err != nil {
		if errors.Is(err, lessflags.ErrHelp) {
			return nil
		}
		if strings.Contains(err.Error(), "unrecognized flag") {
			return fmt.Errorf("%w; see --help", err)
		}
		return err
	}
	if color && noColor {
		return fmt.Errorf("--color and --no-color cannot be specified together")
	}
	if len(remaining) > 0 {
		return fmt.Errorf("unexpected argument: %s", remaining[0])
	}

	sessionID := ""
	if sessionIDFlag != nil {
		sessionID = strings.TrimSpace(*sessionIDFlag)
	}
	if pidFlag != nil && sessionIDFlag != nil {
		return fmt.Errorf("--pid and --session-id cannot be specified together")
	}

	useColor := false
	switch {
	case color:
		useColor = true
	case noColor:
		useColor = false
	default:
		useColor = writerIsTTY(opts.Stdout)
	}

	dirOverride := ""
	dirSet := dirFlag != nil
	if dirSet {
		dirOverride = *dirFlag
	}

	if sessionIDFlag != nil {
		return runModeB(opts, sessionID, dirOverride, dirSet, dryRun, useColor)
	}
	return runModeA(opts, pidFlag, dirOverride, dirSet, dryRun, useColor)
}

func runModeA(opts *Options, pidFlag *int, dirOverride string, dirSet, dryRun, useColor bool) error {
	startPID := opts.PID
	if pidFlag != nil {
		startPID = *pidFlag
	}

	resolveOpts := procresolve.Options{
		GrokHome:  opts.GrokHome,
		ListProcs: opts.ListProcs,
		Lsof:      opts.Lsof,
	}
	anc, ok := procresolve.FindAncestorGrok(startPID, resolveOpts)
	result, err := procresolve.ResolveFromAncestors(startPID, resolveOpts)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("no ancestor grok")
	}
	if result == nil || strings.TrimSpace(result.SessionID) == "" {
		return fmt.Errorf("session not resolved")
	}
	id := result.SessionID

	cwd, err := resolveSessionCWD(opts.GrokHome, id, dirOverride, dirSet)
	if err != nil {
		return err
	}

	exe, err := opts.Executable()
	if err != nil {
		return fmt.Errorf("executable: %w", err)
	}
	followUp := modeAFollowUp(exe, id, opts.Env, cwd, dirSet)

	if dryRun {
		writeModeAPlan(opts.Stdout, useColor, anc.PID, id, cwd, followUp)
		return nil
	}
	if err := opts.OpenInNewTerminal(cwd, followUp); err != nil {
		return err
	}
	opened := paint("Opened", ansiGreen, useColor)
	fmt.Fprintf(opts.Stdout, "%s new window; launching grok-fork --session-id %s\n", opened, id)
	return nil
}

func runModeB(opts *Options, sessionID, dirOverride string, dirSet, dryRun, useColor bool) error {
	if sessionID == "" {
		return fmt.Errorf("session id is required")
	}
	cwd, err := resolveSessionCWD(opts.GrokHome, sessionID, dirOverride, dirSet)
	if err != nil {
		return err
	}
	bin, err := ResolveGrokBin(opts.GrokBin, opts.LookPath)
	if err != nil {
		return err
	}
	launch := SessionLaunch{
		SessionID: sessionID,
		Dir:       cwd,
		Bin:       bin,
		Env:       opts.Env,
	}
	if dryRun {
		writeModeBPlan(opts.Stdout, useColor, launch)
		return nil
	}
	return launch.Run(opts.RunForeground)
}

func resolveSessionCWD(grokHome, sessionID, dirOverride string, dirSet bool) (string, error) {
	info, err := groksessions.Info(grokHome, sessionID)
	if err != nil {
		return "", err
	}
	cwd := ""
	if dirSet {
		cwd = strings.TrimSpace(dirOverride)
	}
	if cwd == "" {
		cwd = strings.TrimSpace(info.CWD)
	}
	if cwd == "" {
		return "", fmt.Errorf("session %s has empty cwd; pass --dir", sessionID)
	}
	return absDir(cwd)
}

func modeAFollowUp(exe, sessionID string, env []string, dirOverride string, dirSet bool) string {
	var b strings.Builder
	if home, ok := envValue(env, "GROK_HOME"); ok {
		b.WriteString("GROK_HOME=")
		b.WriteString(shell.ShellQuote(home))
		b.WriteByte(' ')
	}
	b.WriteString(shell.ShellQuote(exe))
	b.WriteString(" --session-id ")
	b.WriteString(sessionID)
	if dirSet {
		b.WriteString(" --dir ")
		b.WriteString(shell.ShellQuote(dirOverride))
	}
	return b.String()
}

func writeModeAPlan(w io.Writer, useColor bool, ancestorPID int, id, cwd, command string) {
	fmt.Fprintln(w, "Would open new iTerm2 window")
	fmt.Fprintf(w, "  %s%d\n", paint("ancestor pid: ", ansiGray, useColor), ancestorPID)
	fmt.Fprintf(w, "  %s%s\n", paint("grok id:      ", ansiGray, useColor), id)
	fmt.Fprintf(w, "  %s%s\n", paint("cwd:          ", ansiGray, useColor), cwd)
	fmt.Fprintf(w, "  %s%s\n", paint("command:      ", ansiGray, useColor), command)
}

func writeModeBPlan(w io.Writer, useColor bool, launch SessionLaunch) {
	fmt.Fprintln(w, "Would fork grok session")
	fmt.Fprintf(w, "  %s%s\n", paint("grok id:   ", ansiGray, useColor), launch.SessionID)
	fmt.Fprintf(w, "  %s%s\n", paint("cwd:       ", ansiGray, useColor), launch.Dir)
	fmt.Fprintf(w, "  %s%s\n", paint("command:   ", ansiGray, useColor), launch.CommandLine())
	fmt.Fprintf(w, "  %s%s\n", paint("terminal:  ", ansiGray, useColor), "current")
}

func applyDefaults(opts *Options) {
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}
	if opts.PID == 0 {
		opts.PID = os.Getpid()
	}
	if opts.ListProcs == nil {
		opts.ListProcs = procresolve.ListLiveProcs
	}
	if opts.Lsof == nil {
		opts.Lsof = procresolve.LiveLsof
	}
	if opts.OpenInNewTerminal == nil {
		opts.OpenInNewTerminal = defaultOpenInNewTerminal
	}
	if opts.LookPath == nil {
		opts.LookPath = exec.LookPath
	}
	if opts.Executable == nil {
		opts.Executable = os.Executable
	}
	if opts.RunForeground == nil {
		stdout, stderr := opts.Stdout, opts.Stderr
		opts.RunForeground = func(bin string, argv []string, dir string, env []string) error {
			return defaultRunForeground(bin, argv, dir, env, stdout, stderr)
		}
	}
}

func defaultOpenInNewTerminal(dir, followUp string) error {
	return iterm2.OpenConfig(dir, &iterm2.Config{
		Mode:             iterm2.ModeForceNew,
		FollowUpCommands: []string{followUp},
		SafeInputIgnore:  true,
	})
}

func defaultRunForeground(bin string, argv []string, dir string, env []string, stdout, stderr io.Writer) error {
	cmd := exec.Command(bin, argv...)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if len(env) > 0 {
		cmd.Env = append(append([]string{}, os.Environ()...), env...)
	}
	return cmd.Run()
}

func absDir(p string) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("workspace dir: %w", err)
	}
	if real, e := filepath.EvalSymlinks(abs); e == nil {
		abs = real
	}
	st, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("workspace dir: %w", err)
	}
	if !st.IsDir() {
		return "", fmt.Errorf("workspace dir: not a directory: %s", abs)
	}
	return abs, nil
}

func envValue(env []string, key string) (string, bool) {
	prefix := key + "="
	found := false
	val := ""
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			val = strings.TrimPrefix(e, prefix)
			found = true
		}
	}
	return val, found
}

func paint(s, code string, on bool) string {
	if !on {
		return s
	}
	return code + s + ansiReset
}

func writerIsTTY(w io.Writer) bool {
	type fdWriter interface {
		Fd() uintptr
	}
	f, ok := w.(fdWriter)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}
