package main

import (
	"os"

	"github.com/xhd2015/agent-pro/pkgs/agentrunbridge"
)

const (
	envAgentRun = "SLACK_LISTEN_AGENT_RUN"
)

func agentRunBinary() string {
	// Empty lets agentrunbridge default to "agent-run".
	return os.Getenv(envAgentRun)
}

type agentOptions struct {
	Runner           string
	RunnerConfigHome string
	// Env is passed to agent-run as -e KEY=VALUE (e.g. SLACK_MSG_SESSION_ID).
	Env []string
}

// runAgentInteractiveOpen launches thread-mode agent-run via interactive open
// (SeaTalk parity). Capture is empty; TTY owns the session.
func runAgentInteractiveOpen(prompt, sessionID string, opts agentOptions) error {
	_, err := agentrunbridge.RunInteractiveOpen(agentrunbridge.InteractiveOpenOpts{
		SessionID:                     sessionID,
		Prompt:                        prompt,
		Binary:                        agentRunBinary(),
		AgentRunner:                   opts.Runner,
		RunnerConfigHome:              opts.RunnerConfigHome,
		Env:                           opts.Env,
		AllowRelocateResumeSessionDir: true,
	})
	return err
}

// runAgentStateless runs a one-shot agent-run with stdout capture for PostMessage.
func runAgentStateless(prompt string, opts agentOptions) (string, error) {
	result, err := agentrunbridge.Run(agentrunbridge.RunOpts{
		Prompt:           prompt,
		Binary:           agentRunBinary(),
		AgentRunner:      opts.Runner,
		RunnerConfigHome: opts.RunnerConfigHome,
		Stateless:        true,
		CaptureStdout:    true,
		WaitReady:        false,
	})
	if err != nil {
		return "", err
	}
	return result.Stdout, nil
}
