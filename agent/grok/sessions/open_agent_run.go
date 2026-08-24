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
	Delivered         bool   // true when the prompt was already delivered (ModeSend)
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
	// Return (nil, nil) for soft miss; error containing "ambiguous" → warning + miss.
	AgentRunOpen func(grokSessionID, prompt string) (*AgentRunOpenResult, error)
}

// preferAgentRunOpen runs the injectable hook or production prefer for --open.
// Soft miss → ok=false (caller falls back to grok --resume).
func preferAgentRunOpen(hooks agentRunOpenHooks, grokSessionID, prompt string) (hit *AgentRunOpenResult, warn string, ok bool) {
	if hooks.NoAgentRun {
		return nil, "", false
	}
	if hooks.AgentRunOpen != nil {
		res, err := hooks.AgentRunOpen(grokSessionID, prompt)
		if err != nil {
			if strings.Contains(err.Error(), "ambiguous") {
				return nil, "warning: " + err.Error() + "; falling back to grok --resume", false
			}
			return nil, "", false
		}
		if res == nil {
			return nil, "", false
		}
		return res, "", true
	}
	// Production only when no L2 inject (keeps fakes off real agent-run home).
	if hooks.SkipProduction {
		return nil, "", false
	}
	return tryAgentRunOpen(hooks, grokSessionID, prompt)
}

func tryAgentRunOpen(hooks agentRunOpenHooks, grokSessionID, prompt string) (hit *AgentRunOpenResult, warn string, ok bool) {
	home, err := resolveAgentRunHome(hooks.AgentRunHome)
	if err != nil || home == "" {
		return nil, "", false
	}
	store, err := agentstorage.NewFileStore(home)
	if err != nil {
		return nil, "", false
	}
	meta, err := agentstorage.FindByGrokSessionID(store, grokSessionID)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "ambiguous") {
			return nil, "warning: " + msg + "; falling back to grok --resume", false
		}
		return nil, "", false
	}
	arID := strings.TrimSpace(meta.SessionID)
	if arID == "" {
		return nil, "", false
	}

	mode, _, found, err := agentrunapi.Classify(store, arID, nil)
	if err != nil || !found {
		return nil, "", false
	}

	cwd, cwdErr := resolveOpenCWD(hooks.Info, hooks.Dir)
	if cwdErr != nil {
		// Resume/run need cwd; live send/focus can proceed without.
		if mode != agentrunapi.ModeSend {
			return nil, "", false
		}
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
				}, "", true
			}
			_, ferr := agentrunapi.FocusSession(agentrunapi.FocusOpts{
				Store:     store,
				SessionID: arID,
			})
			if ferr != nil {
				return nil, "warning: agent-run session is live but focus failed: " + ferr.Error() + "; falling back to grok --resume", false
			}
			return &AgentRunOpenResult{
				AgentRunSessionID: arID,
				Mode:              AgentRunOpenModeFocus,
				Focused:           true,
				CWD:               cwd,
			}, "", true
		}
		if hooks.DryRun {
			return &AgentRunOpenResult{
				AgentRunSessionID: arID,
				Mode:              AgentRunOpenModeSend,
				Delivered:         true,
				CWD:               cwd,
			}, "", true
		}
		stderr := hooks.Stderr
		if stderr == nil {
			stderr = os.Stderr
		}
		err = agentrunapi.AutoSendOrResume(context.Background(), agentrunapi.Opts{
			SessionID: arID,
			Prompt:    prompt,
			Open:      true,
			NoSubmit:  hooks.NoSubmit,
			Store:     store,
			Stdout:    io.Discard, // enqueue id must not leak onto kck/agent-pro send stdout
			Stderr:    stderr,
		})
		if err != nil {
			return nil, "", false
		}
		return &AgentRunOpenResult{
			AgentRunSessionID: arID,
			Mode:              AgentRunOpenModeSend,
			Delivered:         true,
			CWD:               cwd,
		}, "", true

	case agentrunapi.ModeResume, agentrunapi.ModeRun:
		modeStr := AgentRunOpenModeResume
		if mode == agentrunapi.ModeRun {
			modeStr = AgentRunOpenModeRun
		}
		if cwd == "" {
			return nil, "", false
		}
		bin, berr := resolveAgentRunBin(hooks.AgentRunBin, hooks.LookPath)
		if berr != nil {
			return nil, "", false
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
			}, "", true
		}
		openFn := hooks.OpenInNewWindow
		if openFn == nil {
			openFn = defaultOpenInNewWindow
		}
		if err := openFn(cwd, cmdLine); err != nil {
			return nil, "", false
		}
		return &AgentRunOpenResult{
			AgentRunSessionID: arID,
			Mode:              modeStr,
			Opened:            true,
			Delivered:         strings.TrimSpace(prompt) != "",
			Command:           cmdLine,
			CWD:               cwd,
		}, "", true

	default:
		return nil, "", false
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
