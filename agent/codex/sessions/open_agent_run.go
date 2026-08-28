package sessions

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

const (
	// AgentRunOpenModeSend is live follow-up via agent-run send queue.
	AgentRunOpenModeSend = "send"
	// AgentRunOpenModeResume relaunches an exited bound session via agent-run.
	AgentRunOpenModeResume = "resume"
	// AgentRunOpenModeRun starts/rebinds via agent-run run.
	AgentRunOpenModeRun = "run"
	// AgentRunOpenModeFocus focuses a live agent-run host (open with empty prompt).
	AgentRunOpenModeFocus = "focus"
)

// AgentRunOpenResult is a successful prefer-path open/deliver (or dry-run resolve).
type AgentRunOpenResult struct {
	AgentRunSessionID string
	Mode              string // send | resume | run | focus
	Opened            bool   // true when a new window was (or would be) launched
	Focused           bool   // true when an existing agent-run host was focused
	Delivered         bool   // true when the prompt was already delivered
	Command           string // follow-up command for resume/run window (dry-run / ack)
	CWD               string
}

// agentRunOpenHooks are the shared injectables for send/open prefer.
type agentRunOpenHooks struct {
	NoAgentRun   bool
	AgentRunHome string
	DryRun       bool
	NoSubmit     bool
	Dir          string // optional workspace override
	Info         *SessionInfo

	SkipProduction  bool // true when L2 injects ListProcs/ListITerm (keep fakes off real home)
	OpenInNewWindow func(dir, followUp string) error
	AgentRunBin     string
	LookPath        func(file string) (string, error)
	Stderr          io.Writer

	// AgentRunOpen, when set, replaces production lookup+deliver.
	// Return (nil, nil) for soft miss (not managed).
	// Return (nil, err) with "ambiguous" → warning + soft miss.
	// Return (nil, err) otherwise → hard error (managed deliver failed).
	// Return (hit, nil) on success.
	AgentRunOpen func(codexSessionID, prompt string) (*AgentRunOpenResult, error)
}

// preferAgentRunOpen runs the injectable hook or production prefer.
// Soft miss: hit==nil && err==nil (caller may fall through to iTerm).
// Hard fail: err!=nil (managed id; never fall through to bare codex --resume).
// Success: hit!=nil && err==nil.
func preferAgentRunOpen(hooks agentRunOpenHooks, codexSessionID, prompt string) (hit *AgentRunOpenResult, warn string, err error) {
	if hooks.NoAgentRun {
		return nil, "", nil
	}
	if hooks.AgentRunOpen != nil {
		res, herr := hooks.AgentRunOpen(codexSessionID, prompt)
		if herr != nil {
			if strings.Contains(herr.Error(), "ambiguous") {
				return nil, "warning: " + herr.Error() + "; falling back to codex resume", nil
			}
			return nil, "", fmt.Errorf("agent-run deliver failed: %w", herr)
		}
		if res == nil {
			return nil, "", nil
		}
		return res, "", nil
	}
	// Production only when no L2 inject (keeps fakes off real agent-run home).
	if hooks.SkipProduction {
		return nil, "", nil
	}
	return tryAgentRunOpen(hooks, codexSessionID, prompt)
}

func tryAgentRunOpen(hooks agentRunOpenHooks, codexSessionID, prompt string) (hit *AgentRunOpenResult, warn string, err error) {
	home, herr := resolveAgentRunHome(hooks.AgentRunHome)
	if herr != nil || home == "" {
		return nil, "", nil
	}
	store, serr := agentstorage.NewFileStore(home)
	if serr != nil {
		return nil, "", nil
	}
	meta, ferr := agentstorage.FindByCodexSessionID(store, codexSessionID)
	if ferr != nil {
		msg := ferr.Error()
		if strings.Contains(msg, "ambiguous") {
			return nil, "warning: " + msg + "; falling back to codex resume", nil
		}
		return nil, "", nil // not managed
	}
	arID := strings.TrimSpace(meta.SessionID)
	if arID == "" {
		return nil, "", nil
	}

	mode, _, found, cerr := agentrunapi.Classify(store, arID, nil)
	if cerr != nil {
		return nil, "", fmt.Errorf("agent-run deliver failed for %s: classify: %w", arID, cerr)
	}
	if !found {
		return nil, "", fmt.Errorf("agent-run deliver failed for %s: session disappeared after lookup", arID)
	}

	cwd, cwdErr := resolveOpenCWD(hooks.Info, hooks.Dir)
	if cwdErr != nil {
		cwd = ""
	}

	switch mode {
	case agentrunapi.ModeSend:
		if strings.TrimSpace(prompt) == "" {
			// open-only: focus the live agent-run host.
			if hooks.DryRun {
				return &AgentRunOpenResult{
					AgentRunSessionID: arID,
					Mode:              AgentRunOpenModeFocus,
					Focused:           true,
					CWD:               cwd,
				}, "", nil
			}
			_, ferr := agentrunapi.FocusSession(agentrunapi.FocusOpts{
				Store:     store,
				SessionID: arID,
			})
			if ferr != nil {
				return nil, "", fmt.Errorf("agent-run deliver failed for %s: focus: %w", arID, ferr)
			}
			return &AgentRunOpenResult{
				AgentRunSessionID: arID,
				Mode:              AgentRunOpenModeFocus,
				Focused:           true,
				CWD:               cwd,
			}, "", nil
		}
		if hooks.DryRun {
			return &AgentRunOpenResult{
				AgentRunSessionID: arID,
				Mode:              AgentRunOpenModeSend,
				Delivered:         true,
				CWD:               cwd,
			}, "", nil
		}
		stderr := hooks.Stderr
		if stderr == nil {
			stderr = os.Stderr
		}
		if asrErr := agentrunapi.AutoSendOrResume(context.Background(), agentrunapi.Opts{
			SessionID: arID,
			Prompt:    prompt,
			Open:      true,
			NoSubmit:  hooks.NoSubmit,
			Store:     store,
			Stdout:    io.Discard,
			Stderr:    stderr,
		}); asrErr != nil {
			return nil, "", fmt.Errorf("agent-run deliver failed for %s: %w", arID, asrErr)
		}
		return &AgentRunOpenResult{
			AgentRunSessionID: arID,
			Mode:              AgentRunOpenModeSend,
			Delivered:         true,
			CWD:               cwd,
		}, "", nil

	case agentrunapi.ModeResume, agentrunapi.ModeRun:
		modeStr := AgentRunOpenModeResume
		if mode == agentrunapi.ModeRun {
			modeStr = AgentRunOpenModeRun
		}
		// Exited/unbound: ForceNew with agent-run (never bare codex --resume).
		if cwd == "" {
			if w := strings.TrimSpace(meta.Workspace); w != "" {
				cwd = w
			}
		}
		if cwd == "" {
			return nil, "", fmt.Errorf("agent-run deliver failed for %s: empty workspace; pass --dir", arID)
		}
		bin, berr := resolveAgentRunBin(hooks.AgentRunBin, hooks.LookPath)
		if berr != nil {
			return nil, "", fmt.Errorf("agent-run deliver failed for %s: %w", arID, berr)
		}
		cmdLine := buildAgentRunAutoOpenCommand(bin, arID, cwd, prompt, hooks.NoSubmit)
		if hooks.DryRun {
			return &AgentRunOpenResult{
				AgentRunSessionID: arID,
				Mode:              modeStr,
				Opened:            true,
				Delivered:         strings.TrimSpace(prompt) != "",
				Command:           cmdLine,
				CWD:               cwd,
			}, "", nil
		}
		openFn := hooks.OpenInNewWindow
		if openFn == nil {
			openFn = defaultOpenInNewWindow
		}
		if oerr := openFn(cwd, cmdLine); oerr != nil {
			return nil, "", fmt.Errorf("agent-run deliver failed for %s: open window: %w", arID, oerr)
		}
		return &AgentRunOpenResult{
			AgentRunSessionID: arID,
			Mode:              modeStr,
			Opened:            true,
			Delivered:         strings.TrimSpace(prompt) != "",
			Command:           cmdLine,
			CWD:               cwd,
		}, "", nil

	default:
		return nil, "", fmt.Errorf("agent-run deliver failed for %s: unknown mode %q", arID, mode)
	}
}

func resolveAgentRunBin(explicit string, lookPath func(file string) (string, error)) (string, error) {
	if strings.TrimSpace(explicit) != "" {
		return explicit, nil
	}
	if lookPath == nil {
		lookPath = exec.LookPath
	}
	bin, err := lookPath("agent-run")
	if err != nil {
		return "", fmt.Errorf("agent-run not found on PATH: %w", err)
	}
	return bin, nil
}

// buildAgentRunAutoOpenCommand builds:
//
//	agent-run run --session-id <ar> --auto-send-or-resume --open [--dir DIR] [--no-submit] [-- <prompt>]
func buildAgentRunAutoOpenCommand(bin, arSessionID, dir, prompt string, noSubmit bool) string {
	argv := []string{
		"run",
		"--session-id", arSessionID,
		"--auto-send-or-resume",
		"--open",
	}
	if strings.TrimSpace(dir) != "" {
		argv = append(argv, "--dir", dir)
	}
	if noSubmit {
		argv = append(argv, "--no-submit")
	}
	if p := strings.TrimSpace(prompt); p != "" {
		argv = append(argv, "--", p)
	}
	return quotedForkCommandLine(bin, argv)
}

func writeAgentRunOpenWarn(stderr io.Writer, warn string) {
	if warn == "" {
		return
	}
	if stderr == nil {
		stderr = os.Stderr
	}
	fmt.Fprintln(stderr, warn)
}
