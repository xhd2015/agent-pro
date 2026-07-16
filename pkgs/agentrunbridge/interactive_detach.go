package agentrunbridge

import (
	"strings"
)

// RunInteractiveDetach fills the interactive detach RunOpts profile and calls Run.
//
// Defaults filled:
//
//	AutoSendOrResume, Detach, WaitReady = true
//	NewTerminal, Open, CaptureStdout = false
//	AgentRunner = firstNonEmpty(opts.AgentRunner, "grok-tty")
//
// Uses the same InteractiveOpenOpts input shape as RunInteractiveOpen.
func RunInteractiveDetach(opts InteractiveOpenOpts) (RunResult, error) {
	runner := strings.TrimSpace(opts.AgentRunner)
	if runner == "" {
		runner = "grok-tty"
	}
	return Run(RunOpts{
		Prompt:                        opts.Prompt,
		SessionID:                     opts.SessionID,
		Binary:                        opts.Binary,
		AgentRunner:                   runner,
		RunnerConfigHome:              opts.RunnerConfigHome,
		WorkspaceDir:                  opts.WorkspaceDir,
		NoSubmit:                      opts.NoSubmit,
		AllowRelocateResumeSessionDir: opts.AllowRelocateResumeSessionDir,
		AutoSendOrResume:              true,
		NewTerminal:                   false,
		Open:                          false,
		Detach:                        true,
		WaitReady:                     true,
		CaptureStdout:                 false,
		Env:                           opts.Env,
		ReadyTimeout:                  opts.ReadyTimeout,
		ReadyPollInterval:             opts.ReadyPollInterval,
		LookPath:                      opts.LookPath,
		RunCommand:                    opts.RunCommand,
		RunOutput:                     opts.RunOutput,
		Logf:                          opts.Logf,
	})
}
