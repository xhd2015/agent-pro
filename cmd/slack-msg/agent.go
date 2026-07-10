package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	envAgentRun = "SLACK_LISTEN_AGENT_RUN"
)

func agentRunBinary() string {
	if p := os.Getenv(envAgentRun); p != "" {
		return p
	}
	return "agent-run"
}

type agentOptions struct {
	Runner           string
	RunnerConfigHome string
	SessionMode      string
	SessionID        string
	IsFollowUp       bool
}

func runAgent(prompt string, opts agentOptions) (string, error) {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return "", fmt.Errorf("empty prompt")
	}

	binary := agentRunBinary()
	var args []string

	switch {
	case opts.SessionMode == "stateless":
		args = []string{"run"}
		if opts.Runner != "" {
			args = append(args, "--agent-runner", opts.Runner)
		}
		if opts.RunnerConfigHome != "" {
			args = append(args, "--agent-runner-config-home", opts.RunnerConfigHome)
		}
		args = append(args, prompt)
	case opts.IsFollowUp:
		args = []string{"send", opts.SessionID, prompt}
	default:
		args = []string{"run", "--keep-tty", "--session", opts.SessionID}
		if opts.Runner != "" {
			args = append(args, "--agent-runner", opts.Runner)
		}
		if opts.RunnerConfigHome != "" {
			args = append(args, "--agent-runner-config-home", opts.RunnerConfigHome)
		}
		args = append(args, prompt)
	}

	cmd := exec.Command(binary, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("agent-run failed: %s", msg)
	}
	return strings.TrimSpace(stdout.String()), nil
}
