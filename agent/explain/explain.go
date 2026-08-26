package explain

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	claudeagent "github.com/xhd2015/agent-pro/agent/cli/claude"
	codexagent "github.com/xhd2015/agent-pro/agent/cli/codex"
	commandcodeagent "github.com/xhd2015/agent-pro/agent/cli/commandcode"
	crushagent "github.com/xhd2015/agent-pro/agent/cli/crush"
	grokagent "github.com/xhd2015/agent-pro/agent/cli/grok"
	opencodeagent "github.com/xhd2015/agent-pro/agent/cli/opencode"
	piagent "github.com/xhd2015/agent-pro/agent/cli/pi"
	agentprovider "github.com/xhd2015/agent-pro/agent/cli/provider"
	"github.com/xhd2015/agent-pro/agent/cli/registry"
	agentexec "github.com/xhd2015/agent-pro/agent/exec"
	"github.com/xhd2015/agent-pro/agent/opencode/models"
	"github.com/xhd2015/dot-pkgs/go-pkgs/terminal/color"
	"github.com/xhd2015/less-gen/flags"
)

var explainHelp = `Usage: explain [options] <message> [follow-up messages...]
       explain list [--limit N] [--grep PATTERN]... [--or|--and] [--color]
       explain --set-config [preference flags...]
       explain --show-config

Ask an AI agent a question and get an answer. Sessions are reused when the
positional arguments match a prefix of a previous session's user messages.

Subcommands:
  list        List recent explain sessions (newest first)

Options:
  --model MODEL
              Model to use for generation (or persist with --set-config)
  --agent-runner RUNNER
              Agent runner to use (opencode, codex, grok, or commandcode;
              default: config.json agent_runner, else opencode)
  --set-config
              Persist preference flags into config.json (write-only; no ask)
  --show-config
              Pretty-print explain config.json (missing file prints {})
  --no-config
              Do not read config.json for this run
  --clear-agent-runner
              With --set-config: remove persisted agent_runner
  --clear-model
              With --set-config: remove persisted model
  --color
              Force ANSI color on (even when stderr is not a TTY)
  --no-color
              Force ANSI color off
  -v, --verbose
              Show details about session creation/reuse
  -h, --help  Show this help message

Config file: ~/.agent-pro/dedicated-agents/explain/config.json
Precedence: CLI flag > config.json (unless --no-config) > built-in default
When agent-runner is taken from config, a notice: line is printed on stderr.

Examples:
  explain --set-config --agent-runner codex
  explain --show-config
  explain --no-config "what is this?"
`

var supportedAgentRunners = map[string]bool{
	"opencode":    true,
	"codex":       true,
	"grok":        true,
	"commandcode": true,
}

type Runner interface {
	Start(ctx context.Context, model string, prompt string) (sessionID string, output string, err error)
	Resume(ctx context.Context, model string, prompt string, meta json.RawMessage) (output string, err error)
}

type Runtime struct {
	AgentRunner string
}

func (r *Runtime) Start(ctx context.Context, model string, prompt string) (string, string, error) {
	return r.ask(ctx, model, prompt, "")
}

func (r *Runtime) Resume(ctx context.Context, model string, prompt string, meta json.RawMessage) (string, error) {
	sessionID, err := decodeRunnerSessionID(r.AgentRunner, meta)
	if err != nil {
		return "", err
	}
	_, output, err := r.ask(ctx, model, prompt, sessionID)
	return output, err
}

func (r *Runtime) ask(ctx context.Context, model, prompt, sessionID string) (string, string, error) {
	runnerID := strings.TrimSpace(r.AgentRunner)
	if runnerID == "" {
		runnerID = "opencode"
	}

	// EXPLAIN_AGENT_PATH overrides the selected runner binary. Apply via the
	// registry env key before Build so path lookup succeeds without PATH.
	restore := applyExplainAgentPathEnv(runnerID, os.Getenv("EXPLAIN_AGENT_PATH"))
	defer restore()

	env := agentexec.NewEnv(&agentexec.PathsConfig{
		RootDirName: ".agent-pro",
		DataDirName: "data",
		BinDirName:  "bin",
	}, "AGENT_PRO_CONFIG_HOME")

	built, err := agentprovider.Build(registry.AgentRunnerID(runnerID), "", ".", env)
	if err != nil {
		return "", "", err
	}
	if p := strings.TrimSpace(os.Getenv("EXPLAIN_AGENT_PATH")); p != "" {
		overrideAgentPath(built.Agent, p)
	}

	answer, err := built.Agent.Ask(ctx, prompt, &registry.AskOptions{
		Model:     model,
		Workspace: ".",
		SessionID: sessionID,
	}, nil)
	if err != nil {
		return "", "", err
	}
	newSessionID := runnerSessionIDFromAgent(built.Agent)
	if newSessionID == "" {
		newSessionID = sessionID
	}
	return newSessionID, answer, nil
}

func runnerEnvKey(runnerID string) string {
	switch runnerID {
	case "opencode":
		return registry.EnvOpencodeCLIPath
	case "codex":
		return registry.EnvCodexCLIPath
	case "grok":
		return registry.EnvGrokCLIPath
	case "commandcode":
		return registry.EnvCommandcodeCLIPath
	default:
		return ""
	}
}

func applyExplainAgentPathEnv(runnerID, path string) (restore func()) {
	path = strings.TrimSpace(path)
	key := runnerEnvKey(runnerID)
	if path == "" || key == "" {
		return func() {}
	}
	prev, had := os.LookupEnv(key)
	_ = os.Setenv(key, path)
	return func() {
		if had {
			_ = os.Setenv(key, prev)
		} else {
			_ = os.Unsetenv(key)
		}
	}
}

func overrideAgentPath(agent registry.Agent, path string) {
	path = strings.TrimSpace(path)
	if path == "" {
		return
	}
	switch a := agent.(type) {
	case *opencodeagent.OpencodeAgent:
		a.AgentPath = path
	case *codexagent.CodexAgent:
		a.AgentPath = path
	case *grokagent.GrokAgent:
		a.AgentPath = path
	case *commandcodeagent.CommandcodeAgent:
		a.AgentPath = path
	case *crushagent.CrushAgent:
		a.AgentPath = path
	case *piagent.PiAgent:
		a.AgentPath = path
	case *claudeagent.ClaudeAgent:
		a.AgentPath = path
	}
}

func runnerSessionIDFromAgent(agent registry.Agent) string {
	switch a := agent.(type) {
	case *opencodeagent.OpencodeAgent:
		return strings.TrimSpace(a.LastSessionID)
	case *codexagent.CodexAgent:
		return strings.TrimSpace(a.LastSessionID)
	case *grokagent.GrokAgent:
		return strings.TrimSpace(a.LastSessionID)
	case *commandcodeagent.CommandcodeAgent:
		return strings.TrimSpace(a.LastSessionID)
	case *crushagent.CrushAgent:
		return strings.TrimSpace(a.LastSessionID)
	case *piagent.PiAgent:
		return strings.TrimSpace(a.LastSessionID)
	default:
		return ""
	}
}

func encodeRunnerMeta(runner, sessionID string) json.RawMessage {
	sessionID = strings.TrimSpace(sessionID)
	switch runner {
	case "codex":
		return mustMarshalJSON(map[string]string{"codex_thread_id": sessionID})
	default:
		return mustMarshalJSON(map[string]string{"session_id": sessionID})
	}
}

func decodeRunnerSessionID(runner string, meta json.RawMessage) (string, error) {
	if len(meta) == 0 {
		return "", fmt.Errorf("empty %s runner meta", runner)
	}
	var fields map[string]string
	if err := json.Unmarshal(meta, &fields); err != nil {
		return "", fmt.Errorf("parse %s meta: %w", runner, err)
	}
	if id := strings.TrimSpace(fields["session_id"]); id != "" {
		return id, nil
	}
	if runner == "codex" {
		if id := strings.TrimSpace(fields["codex_thread_id"]); id != "" {
			return id, nil
		}
	}
	return "", fmt.Errorf("session_id not found in %s meta", runner)
}

func RunExplain(args []string) error {
	return RunExplainWithRunner(args, &Runtime{})
}

func RunExplainWithRunner(rawArgs []string, runner Runner) error {
	// Dispatch `list` before global flag parse so `explain list --limit N` works
	// (global parse would reject --limit as unrecognized).
	if len(rawArgs) > 0 && rawArgs[0] == "list" {
		return RunList(rawArgs[1:])
	}

	var model string
	var agentRunner string
	var verbose bool
	var setConfig bool
	var showConfig bool
	var noConfig bool
	var clearAgentRunner bool
	var clearModel bool
	var colorFlag bool
	var noColorFlag bool
	remainArgs, err := flags.
		String("--model", &model).
		String("--agent-runner", &agentRunner).
		Bool("--set-config", &setConfig).
		Bool("--show-config", &showConfig).
		Bool("--no-config", &noConfig).
		Bool("--clear-agent-runner", &clearAgentRunner).
		Bool("--clear-model", &clearModel).
		Bool("--color", &colorFlag).
		Bool("--no-color", &noColorFlag).
		Bool("-v,--verbose", &verbose).
		Help("-h,--help", explainHelp).
		Parse(rawArgs)
	if err != nil {
		return err
	}

	colorMode, err := color.ModeFromFlags(colorFlag, noColorFlag)
	if err != nil {
		return err
	}

	if setConfig && showConfig {
		return fmt.Errorf("--set-config is mutually exclusive with --show-config")
	}
	if setConfig && noConfig {
		return fmt.Errorf("--set-config is mutually exclusive with --no-config")
	}
	if showConfig && noConfig {
		return fmt.Errorf("--show-config is mutually exclusive with --no-config")
	}
	if (clearAgentRunner || clearModel) && !setConfig {
		return fmt.Errorf("--clear-agent-runner/--clear-model require --set-config")
	}

	if showConfig {
		if verbose || clearAgentRunner || clearModel || colorFlag || noColorFlag || model != "" || agentRunner != "" || len(remainArgs) > 0 {
			return fmt.Errorf("--show-config does not take a message or preference flags")
		}
		return runShowConfig()
	}

	if setConfig {
		if verbose || colorFlag || noColorFlag || len(remainArgs) > 0 {
			return fmt.Errorf("--set-config does not take a message")
		}
		return runSetConfig(setConfigPrefs{
			agentRunner:      agentRunner,
			model:            model,
			clearAgentRunner: clearAgentRunner,
			clearModel:       clearModel,
		})
	}

	cliModel := strings.TrimSpace(model)
	cliAgentRunner := strings.TrimSpace(agentRunner)
	agentRunnerFromConfig := false

	if !noConfig {
		cfg, loadErr := loadConfig()
		if loadErr != nil {
			return loadErr
		}
		if cliAgentRunner == "" {
			if fromCfg := strings.TrimSpace(cfg.AgentRunner); fromCfg != "" {
				cliAgentRunner = fromCfg
				agentRunnerFromConfig = true
			}
		}
		if cliModel == "" {
			cliModel = strings.TrimSpace(cfg.Model)
		}
	}

	agentRunner = cliAgentRunner
	model = cliModel

	if agentRunner == "" {
		agentRunner = "opencode"
	}
	if !supportedAgentRunners[agentRunner] {
		return fmt.Errorf("unsupported agent runner: %s (supported: opencode, codex, grok, commandcode)", agentRunner)
	}

	if len(remainArgs) == 0 {
		return fmt.Errorf("missing message argument\n\n%s", explainHelp)
	}

	if agentRunnerFromConfig {
		printAgentRunnerFromConfigNotice(os.Stderr, colorMode, agentRunner)
	}

	firstMsg := remainArgs[0]
	var followUp string
	if len(remainArgs) >= 2 {
		followUp = strings.Join(remainArgs[1:], " ")
	}

	if r, ok := runner.(*Runtime); ok {
		r.AgentRunner = agentRunner
	}

	ctx := context.Background()

	if model == "" {
		_, preferredModel, listErr := models.ListFree()
		if listErr == nil && preferredModel != "" {
			model = preferredModel
		}
	}

	match, err := findMatchingSession(remainArgs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to search sessions: %v\n", err)
	}

	if match != nil {
		agentRunnersMeta := match.Data.AgentRunnersMeta
		if agentRunnersMeta == nil {
			agentRunnersMeta = make(RunnerMeta)
		}
		runnerMetaBytes, ok := agentRunnersMeta[agentRunner]
		if !ok || len(runnerMetaBytes) == 0 {
			return fmt.Errorf("session found but no %s runner meta available", agentRunner)
		}

		prompt := firstMsg
		if followUp != "" {
			followUpArgs := remainArgs[match.MatchedCount:]
			prompt = strings.Join(followUpArgs, " ")
		}

		if verbose {
			matchedMsgs := userMessageSlice(match.Data)[:match.MatchedCount]
			fmt.Fprintf(os.Stderr, "[explain] matched session %s (%d msg prefix: %v)\n", filepath.Base(match.SessionDir), match.MatchedCount, matchedMsgs)
			fmt.Fprintf(os.Stderr, "[explain] resuming %s session...\n", agentRunner)
		}

		output, resumeErr := runner.Resume(ctx, model, prompt, runnerMetaBytes)
		if resumeErr != nil {
			return fmt.Errorf("resume session failed: %w", resumeErr)
		}
		if output == "" {
			return fmt.Errorf("agent returned empty response")
		}

		if verbose {
			fmt.Fprintf(os.Stderr, "[explain] done\n")
		}

		match.Data.Messages = append(match.Data.Messages, Message{Role: "user", Message: prompt})
		match.Data.Messages = append(match.Data.Messages, Message{Role: "assistant", Message: output})
		if model != "" {
			match.Data.Model = model
		}
		if err := updateSession(match.SessionDir, match.Data); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to update session: %v\n", err)
		}

		fmt.Print(output)
		return nil
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "[explain] no matching session\n")
		fmt.Fprintf(os.Stderr, "[explain] starting new %s session...\n", agentRunner)
	}

	sessionID, output, startErr := runner.Start(ctx, model, firstMsg)
	if startErr != nil {
		return fmt.Errorf("start session failed: %w", startErr)
	}
	if output == "" {
		return fmt.Errorf("agent returned empty response")
	}

	runnerMeta := make(RunnerMeta)
	runnerMeta[agentRunner] = encodeRunnerMeta(agentRunner, sessionID)

	data := SessionData{
		AgentRunner:      agentRunner,
		Model:            model,
		AgentRunnersMeta: runnerMeta,
		Messages: []Message{
			{Role: "user", Message: firstMsg},
			{Role: "assistant", Message: output},
		},
	}

	sessionDir, saveErr := saveSession(firstMsg, data)
	if saveErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to save session: %v\n", saveErr)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "[explain] saved to %s\n", filepath.Base(sessionDir))
	}

	if followUp != "" {
		m := &MatchResult{
			SessionDir: sessionDir,
			Data:       data,
		}
		followUpOutput, resumeErr := runner.Resume(ctx, model, followUp, runnerMeta[agentRunner])
		if resumeErr != nil {
			return fmt.Errorf("follow-up resume failed: %w", resumeErr)
		}
		if followUpOutput == "" {
			return fmt.Errorf("agent returned empty response")
		}

		if verbose {
			fmt.Fprintf(os.Stderr, "[explain] follow-up done\n")
		}

		m.Data.Messages = append(m.Data.Messages, Message{Role: "user", Message: followUp})
		m.Data.Messages = append(m.Data.Messages, Message{Role: "assistant", Message: followUpOutput})
		if err := updateSession(sessionDir, m.Data); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to update session: %v\n", saveErr)
		}

		fmt.Print(followUpOutput)
		return nil
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "[explain] done\n")
	}
	fmt.Print(output)
	return nil
}

func mustMarshalJSON(v interface{}) json.RawMessage {
	bytes, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("marshal json: %v", err))
	}
	return json.RawMessage(bytes)
}

// printAgentRunnerFromConfigNotice writes one stderr line when agent-runner
// was resolved from config.json. Only the "notice:" prefix is colored (gray).
func printAgentRunnerFromConfigNotice(w io.Writer, mode color.Mode, runner string) {
	if w == nil {
		w = os.Stderr
	}
	style := color.Style{Enabled: color.EnabledFor(mode, w)}
	fmt.Fprintf(w, "%s agent-runner=%s (from config)\n", style.Gray("notice:"), runner)
}
