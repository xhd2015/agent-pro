package agentui

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	codexagent "github.com/xhd2015/agent-pro/agent/cli/codex"
	crushagent "github.com/xhd2015/agent-pro/agent/cli/crush"
	grokagent "github.com/xhd2015/agent-pro/agent/cli/grok"
	piagent "github.com/xhd2015/agent-pro/agent/cli/pi"
	agentprovider "github.com/xhd2015/agent-pro/agent/cli/provider"
	"github.com/xhd2015/agent-pro/agent/cli/registry"
	"github.com/xhd2015/agent-pro/agent/event/codex_types"
	eventprint "github.com/xhd2015/agent-pro/agent/event/print"
	types "github.com/xhd2015/agent-pro/agent/event/types"
	agentexec "github.com/xhd2015/agent-pro/agent/exec"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/agent-pro/pkgs/groktty"
)

// RunOptions configures a headless agent-run invocation.
type RunOptions struct {
	Prompt       string
	Runner       string
	Model        string
	SessionID    string
	JSON         bool
	Workspace    string
	Store        agentstorage.Store
	Stdout       io.Writer
	Stderr       io.Writer
	StreamPhases bool // web: phased assistant start/update/end; CLI: single message events
}

// Run executes the agent runner, streams output, and persists AgentEvents.
func Run(ctx context.Context, opts RunOptions) error {
	if opts.Store == nil {
		return fmt.Errorf("store is required")
	}
	if strings.TrimSpace(opts.Prompt) == "" {
		return fmt.Errorf("prompt is required")
	}
	runner := strings.TrimSpace(opts.Runner)
	if runner == "" {
		cfg, err := opts.Store.Config()
		if err == nil && strings.TrimSpace(cfg.DefaultAgentRunner) != "" {
			runner = cfg.DefaultAgentRunner
		}
	}
	if runner == "" {
		runner = string(registry.AgentRunnerOpencode)
	}

	sessionID := strings.TrimSpace(opts.SessionID)
	if sessionID == "" {
		sessionID = fmt.Sprintf("sess_%d", os.Getpid())
	}

	runnerSessionID := ""
	if sess, err := opts.Store.GetSession(runner, sessionID); err == nil {
		runnerSessionID = strings.TrimSpace(sess.Meta.RunnerSessionID)
		_ = opts.Store.UpdateSessionStatus(runner, sessionID, "running")
	} else {
		createMeta := agentstorage.SessionMeta{
			Runner:    runner,
			SessionID: sessionID,
			Status:    "running",
			Model:     opts.Model,
		}
		if strings.TrimSpace(opts.Workspace) != "" {
			createMeta.Workspace = opts.Workspace
		}
		_ = opts.Store.CreateSession(runner, sessionID, createMeta)
	}

	priorEvents, _, _ := opts.Store.ReadEvents(runner, sessionID, 0)
	runnerPrompt := BuildContinuationPrompt(priorEvents, opts.Prompt)

	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := opts.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	workspace := opts.Workspace
	if workspace == "" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		workspace = wd
	}

	env := agentexec.NewEnv(&agentexec.PathsConfig{
		RootDirName: ".agent-pro",
		DataDirName: "data",
		BinDirName:  "bin",
	}, "AGENT_PRO_CONFIG_HOME")

	emit := func(ev types.AgentEvent) error {
		if ev.Type == types.ActionMessage && strings.TrimSpace(ev.Role) == "" {
			ev.Role = "assistant"
		}
		if ev.Type == types.ActionMessage && ev.Timestamp == 0 {
			ev.Timestamp = time.Now().UnixMilli()
		}
		if err := opts.Store.AppendEvent(runner, sessionID, ev); err != nil {
			return err
		}
		line, err := json.Marshal(ev)
		if err != nil {
			return err
		}
		if opts.JSON {
			_, err = fmt.Fprintln(stdout, string(line))
			return err
		}
		if formatted := eventprint.FormatAgentEvent(ev); formatted != "" {
			_, err = fmt.Fprintln(stdout, formatted)
		}
		return err
	}

	newRunnerSessionID, runErr := streamRunner(ctx, runner, opts.Store.Home(), workspace, env, runnerPrompt, opts.Model, runnerSessionID, opts.StreamPhases, emit, stderr)
	if strings.TrimSpace(newRunnerSessionID) != "" {
		_ = opts.Store.UpdateSessionRunnerSessionID(runner, sessionID, newRunnerSessionID)
	}
	if err := emit(types.AgentEvent{Type: types.ActionDone}); err != nil {
		return err
	}
	status := "finished"
	if runErr != nil {
		status = "error"
	}
	_ = opts.Store.UpdateSessionStatus(runner, sessionID, status)
	return runErr
}

func streamRunner(ctx context.Context, runner, home, workspace string, env *agentexec.Env, prompt, model, runnerSessionID string, streamPhases bool, emit func(types.AgentEvent) error, stderr io.Writer) (string, error) {
	switch registry.AgentRunnerID(runner) {
	case registry.AgentRunnerGrokTTY:
		return streamGrokTTY(ctx, home, workspace, prompt, model, runnerSessionID, emit, stderr)
	case registry.AgentRunnerCodexTTY:
		return streamCodexTTY(ctx, home, workspace, prompt, model, runnerSessionID, emit, stderr)
	case registry.AgentRunnerFakeCodex, registry.AgentRunnerCodex:
		err := streamCodexLike(ctx, runner, workspace, env, prompt, model, streamPhases, emit, stderr)
		return "", err
	default:
		if _, err := agentprovider.Build(registry.AgentRunnerID(runner), "", workspace, env); err != nil {
			return "", err
		}
		return streamViaRegistry(ctx, runner, workspace, env, prompt, model, runnerSessionID, streamPhases, emit)
	}
}

func streamViaRegistry(ctx context.Context, runner, workspace string, env *agentexec.Env, prompt, model, runnerSessionID string, streamPhases bool, emit func(types.AgentEvent) error) (string, error) {
	built, err := agentprovider.Build(registry.AgentRunnerID(runner), "", workspace, env)
	if err != nil {
		return "", err
	}

	streamID := newAssistantStreamID()
	var buf strings.Builder
	started := false

	_, err = built.Agent.Ask(ctx, prompt, &registry.AskOptions{
		Model:     model,
		Workspace: workspace,
		SessionID: runnerSessionID,
	}, func(delta string) {
		if delta == "" {
			return
		}
		buf.WriteString(delta)
		if !streamPhases {
			return
		}
		if !started {
			_ = emit(types.AgentEvent{
				ID:        streamID,
				Type:      types.ActionMessage,
				Role:      "assistant",
				Phase:     types.PhaseStart,
				Timestamp: time.Now().UnixMilli(),
			})
			started = true
		}
		_ = emit(types.AgentEvent{
			ID:        streamID,
			Type:      types.ActionMessage,
			Role:      "assistant",
			Phase:     types.PhaseUpdate,
			Text:      buf.String(),
			Timestamp: time.Now().UnixMilli(),
		})
	})
	if err != nil {
		return runnerSessionIDFromAgent(built.Agent), err
	}

	text := strings.TrimSpace(buf.String())
	if text == "" {
		return runnerSessionIDFromAgent(built.Agent), nil
	}
	if streamPhases {
		if started {
			if err := emit(types.AgentEvent{
				ID:        streamID,
				Type:      types.ActionMessage,
				Role:      "assistant",
				Phase:     types.PhaseEnd,
				Text:      text,
				Timestamp: time.Now().UnixMilli(),
			}); err != nil {
				return runnerSessionIDFromAgent(built.Agent), err
			}
		} else if err := emitPhasedAssistantMessage(emit, streamID, text); err != nil {
			return runnerSessionIDFromAgent(built.Agent), err
		}
	} else {
		if err := emit(types.AgentEvent{
			Type:      types.ActionMessage,
			Role:      "assistant",
			Text:      text,
			Timestamp: time.Now().UnixMilli(),
		}); err != nil {
			return runnerSessionIDFromAgent(built.Agent), err
		}
	}

	return runnerSessionIDFromAgent(built.Agent), nil
}

func runnerSessionIDFromAgent(agent registry.Agent) string {
	switch a := agent.(type) {
	case *grokagent.GrokAgent:
		return strings.TrimSpace(a.LastSessionID)
	case *piagent.PiAgent:
		return strings.TrimSpace(a.LastSessionID)
	case *crushagent.CrushAgent:
		return strings.TrimSpace(a.LastSessionID)
	}
	return ""
}

func streamGrokTTY(ctx context.Context, home, workspace, prompt, model, runnerSessionID string, emit func(types.AgentEvent) error, stderr io.Writer) (string, error) {
	_, grokSessionID, err := groktty.Run(ctx, groktty.RunOptions{
		Home:            home,
		Workspace:       workspace,
		Prompt:          prompt,
		Model:           model,
		ResumeSessionID: runnerSessionID,
		Stderr:          stderr,
		Emit:            emit,
	})
	if strings.TrimSpace(grokSessionID) != "" {
		return grokSessionID, err
	}
	return runnerSessionID, err
}

func streamCodexTTY(ctx context.Context, home, workspace, prompt, model, runnerSessionID string, emit func(types.AgentEvent) error, stderr io.Writer) (string, error) {
	_, codexSessionID, err := groktty.Run(ctx, groktty.RunOptions{
		Home:            home,
		Workspace:       workspace,
		Prompt:          prompt,
		Model:           model,
		ResumeSessionID: runnerSessionID,
		RunnerID:        string(registry.AgentRunnerCodexTTY),
		StderrPrefix:    string(registry.AgentRunnerCodexTTY),
		RegistryDir:     "codex-tty-registry",
		BannerProvider:  "codex",
		BannerMarkers:   []string{"CODEX_TTY_BANNER"},
		DisableTail:     true,
		Stderr:          stderr,
		Emit:            emit,
	})
	if strings.TrimSpace(codexSessionID) != "" {
		return codexSessionID, err
	}
	return runnerSessionID, err
}

func streamCodexLike(ctx context.Context, runner, workspace string, env *agentexec.Env, prompt, model string, streamPhases bool, emit func(types.AgentEvent) error, stderr io.Writer) error {
	built, err := agentprovider.Build(registry.AgentRunnerID(runner), "", workspace, env)
	if err != nil {
		return err
	}
	agentPath := ""
	if ca, ok := built.Agent.(*codexagent.CodexAgent); ok && ca.AgentPath != "" {
		agentPath = ca.AgentPath
	}
	if agentPath == "" {
		var lookErr error
		agentPath, lookErr = env.LookPath(runner)
		if lookErr != nil {
			return lookErr
		}
	}

	args := []string{"exec", "--json", "--skip-git-repo-check", "--cd", workspace}
	if model != "" {
		args = append(args, "--model", model)
	}
	args = append(args, prompt)

	cmd := env.CommandContext(ctx, agentPath, args...)
	cmd.Dir = workspace
	cmd.Stderr = stderr
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	reader := bufio.NewReader(pipe)
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			line = strings.TrimSpace(line)
			if line != "" && strings.HasPrefix(line, "{") {
				var codexEv codex_types.Event
				if err := json.Unmarshal([]byte(line), &codexEv); err == nil {
					for _, ev := range codex_types.FromCodex([]codex_types.Event{codexEv}, "") {
						if ev.Type == types.ActionMessage {
							if strings.TrimSpace(ev.Role) == "" {
								ev.Role = "assistant"
							}
							if ev.Role == "assistant" && ev.Phase == "" {
								if streamPhases {
									if err := emitPhasedAssistantMessage(emit, "", ev.Text); err != nil {
										return err
									}
								} else if err := emit(ev); err != nil {
									return err
								}
								continue
							}
						}
						if err := emit(ev); err != nil {
							return err
						}
					}
				}
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}
	return cmd.Wait()
}
