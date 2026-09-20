package agentruncli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	flags "github.com/xhd2015/less-flags"
	"github.com/xhd2015/xgo/support/cmd"
	"golang.org/x/term"
)

// dshWebConsumedFlags lists the recorded run flags the browser handoff handles
// itself. Every other parsed run flag is meaningless for a one-shot handoff:
// it prints a stderr notice and is dropped. Unregistered flags never reach
// this point because less-flags rejects them at Parse.
var dshWebConsumedFlags = map[string]bool{
	"--agent-runner": true,
	"--dir":          true,
	"--prompt-file":  true,
	"--browser":      true,
	"--open":         true,
	"--no-submit":    true,
	"-v":             true,
	"--verbose":      true,
}

// dshWebArgv builds a browser handoff, not an agent-run stored or streaming run.
// Prompt files stay file-backed and are resolved against the caller's cwd.
// browser forwards to `dsh web open --browser NAME` when non-empty; DSH owns
// the valid browser names.
func dshWebArgv(prompt, promptFile, dir, browser string, open, noSubmit bool) ([]string, error) {
	if promptFile != "" && prompt != "" {
		return nil, fmt.Errorf("--prompt-file and positional prompt are mutually exclusive")
	}
	if promptFile == "" && strings.TrimSpace(prompt) == "" {
		return nil, fmt.Errorf("dsh-web requires a prompt or --prompt-file")
	}
	if strings.TrimSpace(dir) == "" {
		dir = "."
	}
	workspace, err := resolveRunDir(dir)
	if err != nil {
		return nil, err
	}
	argv := []string{"web", "open", workspace}
	if promptFile != "" {
		path, err := filepath.Abs(promptFile)
		if err != nil {
			return nil, fmt.Errorf("resolve --prompt-file: %w", err)
		}
		argv = append(argv, "--prompt-file", path)
	} else {
		argv = append(argv, "--prompt", prompt)
	}
	if browser != "" {
		argv = append(argv, "--browser", browser)
	}
	if noSubmit {
		argv = append(argv, "--no-submit")
	}
	if !open {
		argv = append(argv, "--no-open")
	}
	return argv, nil
}

// grayNoticePrefix returns the `notice:` prefix, gray-wrapped when color is
// enabled; the notice body stays plain.
func grayNoticePrefix(colorEnabled bool) string {
	if !colorEnabled {
		return "notice:"
	}
	return "\x1b[90mnotice:\x1b[0m"
}

// printDshWebNotices writes one line per recorded run flag the browser handoff
// does not consume, in recorded order, so users learn which flags were dropped:
// `notice: dsh-web ignores --color; it only launches a DSH Web session`.
func printDshWebNotices(w io.Writer, recorded flags.Flags, colorEnabled bool) {
	prefix := grayNoticePrefix(colorEnabled)
	for _, flag := range recorded {
		if dshWebConsumedFlags[flag.Flag] {
			continue
		}
		fmt.Fprintf(w, "%s dsh-web ignores %s; it only launches a DSH Web session\n", prefix, flag.Flag)
	}
}

// dshWebNoticeColorEnabled resolves the notice color auto policy: gray only
// when stderr is a terminal and NO_COLOR is unset or empty.
func dshWebNoticeColorEnabled() bool {
	return term.IsTerminal(int(os.Stderr.Fd())) && os.Getenv("NO_COLOR") == ""
}

// runDSHWeb leaves session identity, input validation, and browser opening to DSH.
// Success means the handoff finished, not that the agent completed its task.
// Recorded flags outside dshWebConsumedFlags print notices and are dropped.
func runDSHWeb(recorded flags.Flags, prompt, promptFile, dir, browser string, open, noSubmit, verbose bool) error {
	argv, err := dshWebArgv(prompt, promptFile, dir, browser, open, noSubmit)
	if err != nil {
		return err
	}
	printDshWebNotices(os.Stderr, recorded, dshWebNoticeColorEnabled())
	command := cmd.New()
	if verbose {
		command.Debug()
	}
	if err := command.Run("dsh", argv...); err != nil {
		return fmt.Errorf("dsh-web launch failed: %w", err)
	}
	return nil
}
