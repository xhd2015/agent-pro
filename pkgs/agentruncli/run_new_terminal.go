package agentruncli

import (
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/agentdriver"
	"github.com/xhd2015/agent-pro/pkgs/shell"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
	flags "github.com/xhd2015/less-flags"
)

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

func quoteFollowUpArgv(tokens []string) string {
	quoted := make([]string, 0, len(tokens))
	for _, tok := range tokens {
		quoted = append(quoted, shell.ShellQuote(tok))
	}
	return strings.Join(quoted, " ")
}

// openInNewTerminalFromRecorded ForceNew-opens iTerm and re-execs this binary
// with recorded flags minus --new-terminal. The launcher does not spawn the provider.
func openInNewTerminalFromRecorded(dir string, recorded flags.Flags, remain []string, promptFile string, eventBus EventBusOpts, sessionID, runner string) error {
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
	tokens, err := agentdriver.MustArgv(resolved, reconstructRunRemainder(recorded, remain, promptFile)...)
	if err != nil {
		return err
	}
	if err := iterm2.OpenConfig(dir, &iterm2.Config{
		Mode:             iterm2.ModeForceNew,
		FollowUpCommands: []string{quoteFollowUpArgv(tokens)},
		SafeInputIgnore:  true,
	}); err != nil {
		return err
	}
	payloadWorkspace := dir
	NotifyOnOpenPath("new-terminal", eventBus, strings.TrimSpace(sessionID), runner, payloadWorkspace)
	return nil
}
