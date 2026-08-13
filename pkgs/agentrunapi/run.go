package agentrunapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentdriver"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/agent-pro/pkgs/shell"
)

// DefaultRunTimeout is used when RunOpts.Timeout is zero.
const DefaultRunTimeout = 30 * time.Minute

// DefaultRunPollInterval is used when RunOpts.PollInterval is zero.
const DefaultRunPollInterval = 2 * time.Second

// RunOpts starts one agent turn and waits until it is done.
//
// Zero values:
//   - OpenTerminal false → detach (no iTerm)
//   - KeepAliveDetached false → /exit after turn-complete (detach only)
//   - ExitOnFinishTerminal false → leave the iTerm up (open only)
//   - Timeout 0 → DefaultRunTimeout (30m)
//
// Store / StoreHome are caller-supplied. Run does not invent a home.
// Empty both → NewFileStore("") (process AGENT_RUN_HOME or ~/.agent-run).
type RunOpts struct {
	Prompt               string
	WorkspaceDir         string
	Model                string
	ModelReasoningEffort string
	AgentRunner          string
	SessionID            string
	OpenTerminal         bool
	KeepAliveDetached    bool
	ExitOnFinishTerminal bool
	Timeout              time.Duration
	PollInterval         time.Duration
	StoreHome            string
	Store                agentstorage.Store
	Env                  []string
	Driver               agentdriver.Driver
	// ResultFile, when set, is an extra done signal: valid JSON at this path
	// ends the wait even if the TTY is still busy. RunJSON sets this.
	ResultFile string

	// Launch replaces production start (tests). Receives normalized opts.
	Launch func(ctx context.Context, opts RunOpts) error
	// Wait replaces production wait-until-done (tests).
	Wait func(ctx context.Context, handle RunHandle) error
	// OpenFn replaces iTerm ForceNew when OpenTerminal is true and Launch is nil.
	OpenFn func(dir, followUp string) error
	// SoftExit replaces production /exit inject (tests).
	SoftExit func(store agentstorage.Store, meta agentstorage.SessionMeta, runner string)
}

// RunJSONOpts is Run plus a JSON example the agent must write.
type RunJSONOpts struct {
	RunOpts
	SchemaExample string
}

// RunHandle is passed to Wait after launch.
type RunHandle struct {
	SessionID    string
	WorkspaceDir string
	SessionDir   string
	ResultFile   string
	Store        agentstorage.Store
}

// RunResult is the session identity after a successful wait.
type RunResult struct {
	SessionID       string
	RunnerSessionID string
	SessionDir      string
	WorkspaceDir    string
}

// Run starts an agent with Prompt in WorkspaceDir and blocks until the turn
// is done (runner exit, resume-ready, idle-after-work, or ResultFile JSON).
func Run(ctx context.Context, opts RunOpts) (*RunResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := normalizeRunOpts(&opts); err != nil {
		return nil, err
	}
	handle, err := startRun(ctx, opts)
	if err != nil {
		return nil, err
	}
	if opts.Wait != nil {
		if err := opts.Wait(ctx, handle); err != nil {
			return nil, err
		}
	} else if err := waitUntilDone(ctx, opts, handle); err != nil {
		return nil, err
	}
	softExitAfterWait(opts, handle)
	return resultFromHandle(handle), nil
}

// RunJSON appends SchemaExample + a temp result path to Prompt, calls Run,
// and returns the file contents. The temp path is absolute and outside WorkspaceDir.
func RunJSON(ctx context.Context, opts RunJSONOpts) (string, error) {
	schema := strings.TrimSpace(opts.SchemaExample)
	if schema == "" {
		return "", fmt.Errorf("schema example is required")
	}
	resultPath, err := newJSONResultPath()
	if err != nil {
		return "", err
	}
	opts.ResultFile = resultPath
	opts.Prompt = appendJSONResultInstructions(opts.Prompt, resultPath, schema)
	if _, err := Run(ctx, opts.RunOpts); err != nil {
		return "", err
	}
	data, err := os.ReadFile(resultPath)
	if err != nil {
		return "", fmt.Errorf("result file %s: %w", resultPath, err)
	}
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || !json.Valid(trimmed) {
		return "", fmt.Errorf("result file %s: not valid JSON", resultPath)
	}
	return string(data), nil
}

func normalizeRunOpts(opts *RunOpts) error {
	if opts == nil {
		return fmt.Errorf("opts is required")
	}
	if strings.TrimSpace(opts.Prompt) == "" {
		return fmt.Errorf("prompt is required")
	}
	if opts.Timeout <= 0 {
		opts.Timeout = DefaultRunTimeout
	}
	if opts.PollInterval <= 0 {
		opts.PollInterval = DefaultRunPollInterval
	}
	if strings.TrimSpace(opts.SessionID) == "" {
		opts.SessionID = fmt.Sprintf("run-%s", time.Now().UTC().Format("20060102T150405.000000000Z"))
	}
	ws := strings.TrimSpace(opts.WorkspaceDir)
	if ws != "" {
		abs, err := filepath.Abs(ws)
		if err != nil {
			return fmt.Errorf("workspace dir: %w", err)
		}
		opts.WorkspaceDir = abs
	}
	if home := strings.TrimSpace(opts.StoreHome); home != "" {
		abs, err := filepath.Abs(home)
		if err != nil {
			return fmt.Errorf("store home: %w", err)
		}
		opts.StoreHome = abs
	}
	return nil
}

func startRun(ctx context.Context, opts RunOpts) (RunHandle, error) {
	store, err := resolveRunStore(opts)
	if err != nil {
		return RunHandle{}, err
	}
	handle := RunHandle{
		SessionID:    opts.SessionID,
		WorkspaceDir: opts.WorkspaceDir,
		ResultFile:   opts.ResultFile,
		Store:        store,
	}
	if store != nil {
		handle.SessionDir = filepath.Join(store.Home(), "sessions", opts.SessionID)
	}
	if opts.Launch != nil {
		if err := opts.Launch(ctx, opts); err != nil {
			return handle, err
		}
		return handle, nil
	}
	if opts.OpenTerminal {
		if err := launchOpen(opts, store); err != nil {
			return handle, err
		}
		return handle, nil
	}
	if err := launchDetach(ctx, opts, store); err != nil {
		return handle, err
	}
	return handle, nil
}

func resolveRunStore(opts RunOpts) (agentstorage.Store, error) {
	if opts.Store != nil {
		return opts.Store, nil
	}
	// Empty StoreHome → NewFileStore("") uses process AGENT_RUN_HOME or ~/.agent-run.
	// Callers that need isolation pass StoreHome; Run does not invent one.
	return agentstorage.NewFileStore(strings.TrimSpace(opts.StoreHome))
}

func launchDetach(ctx context.Context, opts RunOpts, store agentstorage.Store) error {
	if store == nil {
		return fmt.Errorf("store is required for detach launch")
	}
	return AutoSendOrResume(ctx, Opts{
		SessionID:                     opts.SessionID,
		Prompt:                        opts.Prompt,
		WorkspaceDir:                  opts.WorkspaceDir,
		AgentRunner:                   opts.AgentRunner,
		Model:                         opts.Model,
		ModelReasoningEffort:          opts.ModelReasoningEffort,
		Detach:                        true,
		AllowRelocateResumeSessionDir: true,
		Color:                         true,
		Env:                           append([]string(nil), opts.Env...),
		Driver:                        opts.Driver,
		Store:                         store,
	})
}

func launchOpen(opts RunOpts, store agentstorage.Store) error {
	fuOpts := FollowUpOpts{
		Driver:                        opts.Driver,
		SessionID:                     opts.SessionID,
		Prompt:                        opts.Prompt,
		AgentRunner:                   opts.AgentRunner,
		WorkspaceDir:                  opts.WorkspaceDir,
		AllowRelocateResumeSessionDir: true,
		Open:                          true,
		Color:                         true,
		Model:                         opts.Model,
		ModelReasoningEffort:          opts.ModelReasoningEffort,
		Env:                           append([]string(nil), opts.Env...),
	}
	cmd, err := BuildFollowUpCommand(fuOpts)
	if err != nil {
		return err
	}
	if home := storeHomeForFollowUp(opts, store); home != "" {
		cmd = "AGENT_RUN_HOME=" + shell.ShellQuote(home) + " " + cmd
	}
	return OpenInNewTerminal(OpenInNewTerminalOpts{
		WorkspaceDir: opts.WorkspaceDir,
		FollowUp:     cmd,
		OpenTerminal: opts.OpenFn,
	})
}

func storeHomeForFollowUp(opts RunOpts, store agentstorage.Store) string {
	if store != nil {
		if h := strings.TrimSpace(store.Home()); h != "" {
			return h
		}
	}
	return strings.TrimSpace(opts.StoreHome)
}

func resultFromHandle(h RunHandle) *RunResult {
	res := &RunResult{
		SessionID:    h.SessionID,
		SessionDir:   h.SessionDir,
		WorkspaceDir: h.WorkspaceDir,
	}
	if h.Store != nil && h.SessionID != "" {
		if meta, ok, _ := loadRunSessionMeta(h.Store, h.SessionID); ok {
			res.RunnerSessionID = strings.TrimSpace(meta.RunnerSessionID)
			if w := strings.TrimSpace(meta.Workspace); w != "" {
				res.WorkspaceDir = w
			}
		}
	}
	return res
}

func newJSONResultPath() (string, error) {
	f, err := os.CreateTemp("", "agent-run-result-*.json")
	if err != nil {
		return "", fmt.Errorf("create result temp file: %w", err)
	}
	path := f.Name()
	if err := f.Close(); err != nil {
		_ = os.Remove(path)
		return "", err
	}
	// Leave an empty file so the name is reserved. Empty is not valid JSON,
	// so waitUntilDone will not treat it as done.
	return path, nil
}

func appendJSONResultInstructions(prompt, resultPath, schema string) string {
	var b strings.Builder
	b.WriteString(strings.TrimRight(prompt, "\n"))
	b.WriteString("\n\n")
	b.WriteString("When finished, write a single JSON object to:\n  ")
	b.WriteString(resultPath)
	b.WriteString("\nMatch this example (same keys; values are illustrative):\n")
	b.WriteString(schema)
	b.WriteString("\nWrite atomically (tmp + rename). Do not write the result into the worktree.\n")
	return b.String()
}
