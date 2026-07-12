package agentrunbridge

import "strings"

// BuildArgs maps RunOpts to agent-run CLI arguments after the binary name.
// The returned slice always starts with "run".
func BuildArgs(opts RunOpts) []string {
	args := []string{"run"}

	if !opts.Stateless {
		if sid := strings.TrimSpace(opts.SessionID); sid != "" {
			args = append(args, "--session-id="+sid)
		}
	}

	if runner := strings.TrimSpace(opts.AgentRunner); runner != "" {
		args = append(args, "--agent-runner="+runner)
	}
	if home := strings.TrimSpace(opts.RunnerConfigHome); home != "" {
		args = append(args, "--agent-runner-config-home="+home)
	}

	if opts.AutoSendOrResume {
		args = append(args, "--auto-send-or-resume")
	}
	if opts.KeepTTY {
		args = append(args, "--keep-tty")
	}
	if opts.NewTerminal {
		args = append(args, "--new-terminal")
	}

	if dir := strings.TrimSpace(opts.WorkspaceDir); dir != "" {
		args = append(args, "--dir="+dir)
	}
	if opts.NoSubmit {
		args = append(args, "--no-submit")
	}
	if opts.Open {
		args = append(args, "--open")
	}

	prompt := strings.TrimSpace(opts.Prompt)
	if opts.Open {
		// Open profile: prompt after "--" separator (SeaTalk parity).
		args = append(args, "--", prompt)
	} else {
		// Keep-tty / non-open / stateless: prompt as last arg (no required "--").
		args = append(args, prompt)
	}
	return args
}
