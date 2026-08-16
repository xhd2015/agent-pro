package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
	"github.com/xhd2015/agent-pro/pkgs/agentrunbridge"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

const (
	envAgentRun = "SLACK_LISTEN_AGENT_RUN"
)

func agentRunBinary() string {
	// Empty → agentrunapi DriverBinary "" defaults FollowUp to "agent-run".
	return os.Getenv(envAgentRun)
}

type agentOptions struct {
	Runner           string
	RunnerConfigHome string
	// WorkspaceDir is the agent workspace (from session map dir when set).
	WorkspaceDir string
	// Env is passed to agent-run as -e KEY=VALUE (e.g. SLACK_MSG_SESSION_ID).
	Env []string
}

// runAgentInteractiveOpen launches thread-mode agent via agentrunapi open profile
// (AutoSendOrResume + OpenInNewTerminal ForceNew + WaitReady).
// DriverBinary is agentRunBinary() (SLACK_LISTEN_AGENT_RUN); empty → library "agent-run".
// Capture is empty; TTY owns the session.
func runAgentInteractiveOpen(prompt, sessionID string, opts agentOptions) error {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return fmt.Errorf("session id is required")
	}

	runner := strings.TrimSpace(opts.Runner)
	if runner == "" {
		runner = "grok-tty"
	}

	store, err := openAgentStore()
	if err != nil {
		return err
	}

	// Empty DriverBinary → BuildFollowUpCommand defaults to "agent-run".
	driver := agentRunBinary()

	apiOpts := agentrunapi.Opts{
		SessionID:                     sessionID,
		Prompt:                        prompt,
		WorkspaceDir:                  strings.TrimSpace(opts.WorkspaceDir),
		AgentRunner:                   runner,
		RunnerConfigHome:              strings.TrimSpace(opts.RunnerConfigHome),
		Open:                          true,
		NewTerminal:                   true,
		Detach:                        false,
		AllowRelocateResumeSessionDir: true,
		Env:                           append([]string(nil), opts.Env...),
		Store:                         store,
		Stdout:                        os.Stdout,
		Stderr:                        os.Stderr,
	}

	// Production open profile: ForceNew on run/resume; live send stays in-process.
	apiOpts.RunSession = func(ctx context.Context, o agentrunapi.Opts, meta agentstorage.SessionMeta, found bool) error {
		_ = ctx
		_ = meta
		_ = found
		return openInteractiveInNewTerminal(o, driver)
	}
	apiOpts.ResumeSession = func(ctx context.Context, o agentrunapi.Opts, meta agentstorage.SessionMeta) error {
		_ = ctx
		_ = meta
		return openInteractiveInNewTerminal(o, driver)
	}

	if err := agentrunapi.AutoSendOrResume(context.Background(), apiOpts); err != nil {
		return err
	}

	// Parity with prior interactive-open WaitReady=true profile.
	return agentrunapi.WaitReady(agentrunapi.WaitReadyOpts{
		SessionID: sessionID,
		StatusFn:  makeTTYStatusFn(driver, sessionID),
	})
}

// openInteractiveInNewTerminal builds a ForceNew FollowUp that re-invokes
// agent-run (or SLACK_LISTEN_AGENT_RUN) with --auto-send-or-resume --open
// and no --new-terminal (library BuildFollowUpCommand).
func openInteractiveInNewTerminal(o agentrunapi.Opts, driverBinary string) error {
	opts := agentrunapi.OpenInNewTerminalOpts{
		WorkspaceDir: o.WorkspaceDir,
		FollowUpOpts: agentrunapi.FollowUpOpts{
			// DriverBinary empty → "agent-run" (compat fallback).
			DriverBinary:                  driverBinary,
			SessionID:                     o.SessionID,
			Prompt:                        o.Prompt,
			AgentRunner:                   o.AgentRunner,
			WorkspaceDir:                  o.WorkspaceDir,
			NoSubmit:                      o.NoSubmit,
			AllowRelocateResumeSessionDir: o.AllowRelocateResumeSessionDir,
			Open:                          o.Open,
			Detach:                        o.Detach,
			Env:                           o.Env,
		},
	}
	// Tests set SLACK_LISTEN_AGENT_RUN to a mock script. Exec that follow-up
	// in-process: iTerm ForceNew is unavailable in CI / headless runs.
	if strings.TrimSpace(os.Getenv(envAgentRun)) != "" {
		opts.OpenTerminal = execFollowUpInProcess
	}
	return agentrunapi.OpenInNewTerminal(opts)
}

func execFollowUpInProcess(dir, followUp string) error {
	cmd := exec.Command("sh", "-c", followUp)
	if strings.TrimSpace(dir) != "" {
		cmd.Dir = dir
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func makeTTYStatusFn(driverBinary, sessionID string) func() (string, error) {
	return func() (string, error) {
		bin := strings.TrimSpace(driverBinary)
		if bin == "" {
			bin = "agent-run"
		}
		cmd := exec.Command(bin, "tty", "status", sessionID)
		out, err := cmd.CombinedOutput()
		return string(out), err
	}
}

func openAgentStore() (agentstorage.Store, error) {
	home := strings.TrimSpace(os.Getenv("AGENT_RUN_HOME"))
	if home == "" {
		userHome, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("agent store home: %w", err)
		}
		home = filepath.Join(userHome, ".agent-run")
	}
	return agentstorage.NewFileStore(home)
}

// runAgentStateless runs a one-shot agent-run with stdout capture for PostMessage.
// Keeps agentrunbridge for CaptureStdout (agentrunapi has no stateless capture yet).
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
