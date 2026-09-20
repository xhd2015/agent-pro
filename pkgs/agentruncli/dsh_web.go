package agentruncli

import (
	"fmt"
	"path/filepath"
	"strings"

	flags "github.com/xhd2015/less-flags"
	"github.com/xhd2015/xgo/support/cmd"
)

// dshWebArgv builds a browser handoff, not an agent-run stored or streaming run.
// Prompt files stay file-backed and are resolved against the caller's cwd.
func dshWebArgv(recorded flags.Flags, prompt, promptFile, dir string, open, noSubmit bool) ([]string, error) {
	for _, flag := range recorded {
		switch flag.Flag {
		case "--agent-runner", "--dir", "--prompt-file", "--open", "--no-submit", "-v", "--verbose":
		default:
			return nil, fmt.Errorf("dsh-web does not support %s; it only launches a DSH Web session", flag.Flag)
		}
	}
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
	if noSubmit {
		argv = append(argv, "--no-submit")
	}
	if !open {
		argv = append(argv, "--no-open")
	}
	return argv, nil
}

// runDSHWeb leaves session identity, input validation, and browser opening to DSH.
// Success means the handoff finished, not that the agent completed its task.
func runDSHWeb(recorded flags.Flags, prompt, promptFile, dir string, open, noSubmit, verbose bool) error {
	argv, err := dshWebArgv(recorded, prompt, promptFile, dir, open, noSubmit)
	if err != nil {
		return err
	}
	command := cmd.New()
	if verbose {
		command.Debug()
	}
	if err := command.Run("dsh", argv...); err != nil {
		return fmt.Errorf("dsh-web launch failed: %w", err)
	}
	return nil
}
