package agentrunbridge

import (
	"strings"
	"time"
)

// InteractiveOpenOpts is the minimal interactive --open profile.
// RunInteractiveOpen fills assumed RunOpts flags and calls Run only.
type InteractiveOpenOpts struct {
	SessionID, Prompt, WorkspaceDir, Binary, AgentRunner, RunnerConfigHome string
	NoSubmit                                                               bool
	// ReadyTimeout / ReadyPollInterval optional; 0 → Run defaults (60s / 500ms).
	ReadyTimeout      time.Duration
	ReadyPollInterval time.Duration
	LookPath          func(file string) (string, error)
	RunCommand        func(name string, args ...string) error
	RunOutput         func(name string, args ...string) (string, error)
	Logf              func(format string, args ...any)
}

// RunInteractiveOpen fills the interactive open RunOpts profile and calls Run.
//
// Defaults filled:
//
//	AutoSendOrResume, NewTerminal, Open, WaitReady = true
//	CaptureStdout = false
//	AgentRunner = firstNonEmpty(opts.AgentRunner, "grok-tty")
func RunInteractiveOpen(opts InteractiveOpenOpts) (RunResult, error) {
	runner := strings.TrimSpace(opts.AgentRunner)
	if runner == "" {
		runner = "grok-tty"
	}
	return Run(RunOpts{
		Prompt:            opts.Prompt,
		SessionID:         opts.SessionID,
		Binary:            opts.Binary,
		AgentRunner:       runner,
		RunnerConfigHome:  opts.RunnerConfigHome,
		WorkspaceDir:      opts.WorkspaceDir,
		NoSubmit:          opts.NoSubmit,
		AutoSendOrResume:  true,
		NewTerminal:       true,
		Open:              true,
		WaitReady:         true,
		CaptureStdout:     false,
		ReadyTimeout:      opts.ReadyTimeout,
		ReadyPollInterval: opts.ReadyPollInterval,
		LookPath:          opts.LookPath,
		RunCommand:        opts.RunCommand,
		RunOutput:         opts.RunOutput,
		Logf:              opts.Logf,
	})
}
