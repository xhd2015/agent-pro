package sessions

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/shell"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
	"github.com/xhd2015/less-gen/flags"
)

// ForkHelp is the text for `agent-pro grok session fork --help`.
const ForkHelp = `Usage: agent-pro grok session fork (<session-id> | --tab SEL | --tab-index N) [OPTIONS]

Fork a Grok CLI session using the native flag:
  grok --resume <session-id> --fork-session

Does not use agent-run storage (avoids already-mapped import limits).

Session source (exactly one):
  <session-id>          explicit Grok session id
  --tab SEL             1-based tab index, or next|left|right (right ≡ next)
  --tab-index N         0-based tab index in this iTerm window

Options:
  -n,--new-window       open a new iTerm2 window for the fork
                        (alias: --new-terminal)
  --dir DIR             workspace (default: session info.cwd from GROK_HOME)
  --session-id UUID     optional id for the forked Grok session
  --dry-run             print plan only; do not launch
  -h,--help             show help

Tab selectors use the same window/tab discovery as: kool iterm2 window status.
Relative next/left/right do not wrap; edges error.
Without -n/--new-window, the fork runs in the current terminal.
`

// ForkCommandHelpLine is the parent `agent-pro grok session` help row.
const ForkCommandHelpLine = `  fork   …               fork via grok --resume … (--tab / --tab-index / <id>)`

// ForkOpts drives RunFork. Nil hooks use production probes and launchers.
type ForkOpts struct {
	Stdout, Stderr io.Writer
	GrokHome       string
	Env            []string
	GrokBin        string

	// Tab path hooks (nil → production).
	ListFocusProcs   func() []FocusProc
	Lsof             func(int) []string
	ListITerm        func() ([]iterm2.SessionRef, error)
	CurrentSessionID func() string
	ControllingTTY   func() string
	AncestorTTYs     func() []string

	OpenInNewWindow func(dir, followUp string) error
	RunForeground   func(bin string, argv []string, dir string, env []string) error
	LookPath        func(file string) (string, error)
}

// RunFork implements `agent-pro grok session fork`.
// It does not os.Exit and does not print an `Error:` / `agent-pro:` prefix.
func RunFork(args []string, opts *ForkOpts) error {
	if opts == nil {
		opts = &ForkOpts{}
	}
	applyForkDefaults(opts)

	var newWindow bool
	var dryRun bool
	var dir string
	var newSessionID string
	var tabFlag *string
	var tabIndexFlag *int

	stdout := opts.Stdout
	remaining, err := flags.Bool("-n,--new-window,--new-terminal", &newWindow).
		Bool("--dry-run", &dryRun).
		String("--dir", &dir).
		String("--session-id", &newSessionID).
		String("--tab", &tabFlag).
		Int("--tab-index", &tabIndexFlag).
		HelpFunc("-h,--help", func() {
			txt := strings.TrimPrefix(ForkHelp, "\n")
			fmt.Fprint(stdout, txt)
			if !strings.HasSuffix(txt, "\n") {
				fmt.Fprintln(stdout)
			}
		}).
		HelpNoExit().
		Parse(args)
	if err == flags.ErrHelp {
		return nil
	}
	if err != nil {
		return err
	}

	sessionID, tabMeta, err := ResolveSessionSource(remaining, tabFlag, tabIndexFlag, &SessionSourceOpts{
		ListProcs:        opts.ListFocusProcs,
		Lsof:             opts.Lsof,
		ListITerm:        opts.ListITerm,
		CurrentSessionID: opts.CurrentSessionID,
		ControllingTTY:   opts.ControllingTTY,
		AncestorTTYs:     opts.AncestorTTYs,
	})
	if err != nil {
		return err
	}

	info, err := Info(opts.GrokHome, sessionID)
	if err != nil {
		return err
	}

	cwd := strings.TrimSpace(dir)
	if cwd == "" {
		cwd = strings.TrimSpace(info.CWD)
	}
	if cwd == "" {
		return fmt.Errorf("session %s has empty cwd; pass --dir", sessionID)
	}
	abs, absErr := filepath.Abs(cwd)
	if absErr != nil {
		return fmt.Errorf("workspace dir: %w", absErr)
	}
	if real, e := filepath.EvalSymlinks(abs); e == nil {
		abs = real
	}
	st, stErr := os.Stat(abs)
	if stErr != nil {
		return fmt.Errorf("workspace dir: %w", stErr)
	}
	if !st.IsDir() {
		return fmt.Errorf("workspace dir: not a directory: %s", abs)
	}
	cwd = abs

	bin, err := resolveForkGrokBin(opts.GrokBin, opts.LookPath)
	if err != nil {
		return err
	}
	argv := resumeForkArgv(sessionID, newSessionID)
	cmdLine := quotedForkCommandLine(bin, argv)

	if dryRun {
		fmt.Fprintln(stdout, "Would fork grok session")
		if tabMeta != nil {
			fmt.Fprintf(stdout, "  window:     %s\n", tabMeta.WindowID)
			fmt.Fprintf(stdout, "  tab index:  %d\n", tabMeta.TabIndex)
			fmt.Fprintf(stdout, "  tty:        %s\n", tabMeta.TTY)
		}
		fmt.Fprintf(stdout, "  grok id:   %s\n", sessionID)
		fmt.Fprintf(stdout, "  cwd:       %s\n", cwd)
		fmt.Fprintf(stdout, "  command:   %s\n", cmdLine)
		if newWindow {
			fmt.Fprintln(stdout, "  terminal:  new iTerm2 window")
		} else {
			fmt.Fprintln(stdout, "  terminal:  current")
		}
		return nil
	}

	if newWindow {
		if err := opts.OpenInNewWindow(cwd, cmdLine); err != nil {
			return fmt.Errorf("open new window: %w", err)
		}
		fmt.Fprintf(stdout, "Opened new window; forking grok session %s\n", sessionID)
		return nil
	}

	if err := opts.RunForeground(bin, argv, cwd, opts.Env); err != nil {
		return fmt.Errorf("grok fork: %w", err)
	}
	return nil
}

func applyForkDefaults(opts *ForkOpts) {
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}
	if opts.Env == nil {
		opts.Env = os.Environ()
	}
	if opts.LookPath == nil {
		opts.LookPath = exec.LookPath
	}
	if opts.OpenInNewWindow == nil {
		opts.OpenInNewWindow = func(dir, followUp string) error {
			return iterm2.OpenConfig(dir, &iterm2.Config{
				Mode:             iterm2.ModeForceNew,
				FollowUpCommands: []string{followUp},
				SafeInputIgnore:  true,
			})
		}
	}
	if opts.RunForeground == nil {
		stdout, stderr := opts.Stdout, opts.Stderr
		opts.RunForeground = func(bin string, argv []string, dir string, env []string) error {
			cmd := exec.Command(bin, argv...)
			cmd.Dir = dir
			cmd.Stdin = os.Stdin
			cmd.Stdout = stdout
			cmd.Stderr = stderr
			if len(env) > 0 {
				cmd.Env = env
			}
			return cmd.Run()
		}
	}
}

func resolveForkGrokBin(grokBin string, lookPath func(file string) (string, error)) (string, error) {
	if strings.TrimSpace(grokBin) != "" {
		return grokBin, nil
	}
	if lookPath == nil {
		lookPath = exec.LookPath
	}
	bin, err := lookPath("grok")
	if err != nil {
		return "", fmt.Errorf("grok not found on PATH: %w", err)
	}
	return bin, nil
}

func resumeForkArgv(sessionID, newSessionID string) []string {
	argv := []string{"--resume", sessionID, "--fork-session"}
	if sid := strings.TrimSpace(newSessionID); sid != "" {
		argv = append(argv, "--session-id", sid)
	}
	return argv
}

func quotedForkCommandLine(bin string, argv []string) string {
	parts := append([]string{bin}, argv...)
	quoted := make([]string, 0, len(parts))
	for _, p := range parts {
		quoted = append(quoted, shell.ShellQuote(p))
	}
	return strings.Join(quoted, " ")
}
