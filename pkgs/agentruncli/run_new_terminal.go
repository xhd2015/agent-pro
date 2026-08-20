package agentruncli

import (
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/agentdriver"
	"github.com/xhd2015/agent-pro/pkgs/shell"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
	flags "github.com/xhd2015/less-flags"
)

// execInTerminalEnvEnabled reports whether --new-terminal should exec the
// child command so the parent terminal closes when the run exits.
//
// Production default: false (current behavior preserved).
//
// Env AGENT_RUN_EXEC_IN_TERMINAL=1/true/yes/on → enabled.
// CLI flags --exec-in-terminal / --no-exec-in-terminal override the env.
func execInTerminalEnvEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("AGENT_RUN_EXEC_IN_TERMINAL")))
	switch v {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

// resolveExecInTerminalFlags merges --exec-in-terminal / --no-exec-in-terminal
// flag values with the env-var default. Both flags set → error.
// --exec-in-terminal forces true; --no-exec-in-terminal forces false;
// neither → envDefault.
func resolveExecInTerminalFlags(execFlag, noExecFlag, envDefault bool) (bool, error) {
	if execFlag && noExecFlag {
		return false, fmt.Errorf("--exec-in-terminal and --no-exec-in-terminal are mutually exclusive; cannot use both")
	}
	if execFlag {
		return true, nil
	}
	if noExecFlag {
		return false, nil
	}
	return envDefault, nil
}

// withExecPrefix prepends "exec " to cmd when exec is true so the child
// replaces the login shell and the terminal closes on exit.
func withExecPrefix(cmd string, exec bool) string {
	if !exec || strings.TrimSpace(cmd) == "" {
		return cmd
	}
	return "exec " + cmd
}

// reconstructRunRemainder builds child `run` argv from recorded parent flags.
// --new-terminal is stripped so the iTerm child does not loop.
// Positional remain is placed after `--` unless the parent used --prompt-file.
func reconstructRunRemainder(recorded flags.Flags, remain []string, promptFile string) []string {
	out := make([]string, 0, 2+len(recorded)+len(remain))
	out = append(out, "run")
	out = append(out, recorded.Remove("--new-terminal").Reconstruct()...)
	if strings.TrimSpace(promptFile) == "" && len(remain) > 0 {
		if remain[0] != "--" {
			out = append(out, "--")
		}
		out = append(out, remain...)
	}
	return out
}

func withResolvedRunner(childArgs []string, runner string) []string {
	if len(childArgs) == 0 || strings.TrimSpace(runner) == "" {
		return childArgs
	}
	return append(childArgs[:1], append([]string{"--agent-runner", runner}, childArgs[1:]...)...)
}

func quoteFollowUpArgv(tokens []string) string {
	quoted := make([]string, 0, len(tokens))
	for _, tok := range tokens {
		quoted = append(quoted, shell.ShellQuote(tok))
	}
	return strings.Join(quoted, " ")
}

// openInNewTerminalFromRecorded ForceNew-opens iTerm and re-execs this binary
// with recorded flags minus --new-terminal. The launcher does not spawn the provider.
// When execInTerminal is true, the follow-up is prefixed with "exec " so the
// terminal closes when the child run exits.
func openInNewTerminalFromRecorded(dir string, recorded flags.Flags, remain []string, promptFile string, eventBus EventBusOpts, sessionID, runner string, execInTerminal bool) error {
	exe, err := agentRunExecutable()
	if err != nil {
		return err
	}
	host := mergeHostDriver(agentdriver.Driver{})
	if strings.TrimSpace(host.Binary) == "" {
		host = agentdriver.Driver{Binary: exe}
	}
	resolved, err := agentdriver.Resolve(host)
	if err != nil {
		return err
	}
	childArgs := reconstructRunRemainder(recorded, remain, promptFile)
	// Handle consumes --agent-runner as a global option before this command is
	// re-execed. Put the resolved runner back into the child argv so --new-terminal
	// preserves codex-tty rather than silently falling back to grok-tty.
	childArgs = withResolvedRunner(childArgs, runner)
	tokens, err := agentdriver.MustArgv(resolved, childArgs...)
	if err != nil {
		return err
	}
	followUp := withExecPrefix(quoteFollowUpArgv(tokens), execInTerminal)
	if err := iterm2.OpenConfig(dir, &iterm2.Config{
		Mode:             iterm2.ModeForceNew,
		FollowUpCommands: []string{followUp},
		SafeInputIgnore:  true,
	}); err != nil {
		return err
	}
	payloadWorkspace := dir
	NotifyOnOpenPath("new-terminal", eventBus, strings.TrimSpace(sessionID), runner, payloadWorkspace)
	return nil
}
