package run

import (
	"os"
	"strings"

	lessflags "github.com/xhd2015/less-flags"
)

const RunFlagsEnvVar = "LLM_MOCK_RUN_FLAGS"

// RunCommandHelp documents llm-mock run flags and environment variables.
const RunCommandHelp = `
Usage: llm-mock run [--mock-events-preset NAME] [--mock-mcp SPEC] [--log-events FILE] [--log-http FILE] (grok|codex|opencode|commandcode) [agent-args...]

Start a background llm-mock HTTP server, configure an isolated agent home pointing
at the mock, and run grok, codex, opencode, or commandcode in the foreground.
The mock stops when the agent exits.

Options:
  --mock-events-preset NAME
        Named AgentEvent sequence to seed mock genQueue after config exchanges,
        or "list" to print the preset catalog and exit (no agent, no mock server).
  --mock-events-file FILE
        AgentEvent JSONL appended to genQueue (delay_ms / type=sleep honored).
  --mock-mcp SPEC
        Codex only. Repeatable mock-mcp child: name=duration, name=min-max, or name=hang.
  --log-events FILE
        Append standard AgentEvent JSONL for each served mock response.
        FILE must end with .jsonl. Passed to the mock as --agent-events-file.
  --log-http FILE
        Append full HTTP request/response exchange JSONL for each mock HTTP call.
        FILE must end with .jsonl. Passed to the mock as --log-http.
  --agent-runner-config-home PATH
        Agent data directory (grok: GROK_HOME). Default: AGENT_RUNNER_CONFIG_HOME env.

Environment (orchestrator, grok):
  LLM_MOCK_GROK_HOME                      Explicit grok home (default: temp dir)
  LLM_MOCK_RUN_GROK_COMMAND               Replace grok executable (tests/plumbing)
  LLM_MOCK_RUN_GROK_DEBUG=1               Verbose grok orchestrator stderr debug logs

Environment (orchestrator, codex):
  LLM_MOCK_CODEX_HOME                     Explicit codex home (default: temp dir)
  LLM_MOCK_RUN_CODEX_COMMAND              Replace codex executable (tests/plumbing)
  LLM_MOCK_RUN_CODEX_DEBUG=1              Verbose codex orchestrator stderr debug logs
  LLM_MOCK_EXTRA_MCP_TOML_FILE            Append TOML (e.g. [mcp_servers.*]) to generated config.toml
  LLM_MOCK_MCP                            Comma-separated --mock-mcp SPECs (CLI --mock-mcp wins on duplicate names)
  LLM_MOCK_MCP_COMMAND                    Override mock-mcp executable path (tests/plumbing)

Environment (orchestrator, opencode):
  LLM_MOCK_OPENCODE_HOME                  Explicit opencode HOME (default: temp dir)
  LLM_MOCK_OPENCODE_CONFIG_DIR            Explicit opencode config dir (default: temp dir)
  LLM_MOCK_RUN_OPENCODE_COMMAND           Replace opencode executable (tests/plumbing)
  LLM_MOCK_RUN_OPENCODE_DEBUG=1           Verbose opencode orchestrator stderr debug logs

Environment (orchestrator, commandcode):
  LLM_MOCK_RUN_COMMANDCODE_COMMAND        Replace cmd executable (tests/plumbing)
  LLM_MOCK_RUN_COMMANDCODE_DEBUG=1        Verbose commandcode orchestrator stderr debug logs
  HOME is set to an isolated temp dir so ~/.commandcode from real usage is not touched.
  COMMANDCODE_SANDBOX=true and COMMAND_CODE_API_KEY=sk-mock are always injected.
  LLM_MOCK_RUN_FLAGS="--help" ...         Show llm-mock flags (when using llm-mock-run-commandcode shortcut)

Shared environment:
  LLM_MOCK_CONFIG_FILE, LLM_MOCK_CONFIG   Mock exchange config JSON path
  LLM_MOCK_EVENTS_FILE                    Optional input exchanges JSONL (appended)
  LLM_MOCK_RUN_FLAGS                      Space-separated default run flags (prepended before argv; CLI wins on duplicate)
  AGENT_RUNNER_CONFIG_HOME                Agent data directory when --agent-runner-config-home is unset

Config is optional; with no config env vars, uses inline default {"exchanges": []}.

Examples:
  llm-mock run --mock-events-preset=list
  llm-mock run grok
  llm-mock run codex
  llm-mock run commandcode "explain this code"
  llm-mock run --mock-events-preset=think-message grok
  llm-mock run --log-events session.jsonl codex exec -m mock-model "hi"
  llm-mock run --log-http http.jsonl grok
  llm-mock run --log-events session.jsonl --log-http http.jsonl grok -p hello --always-approve
  llm-mock run opencode run "hi" --model llm-mock/mock-model
  llm-mock run --mock-events-preset think-tool-message commandcode "list files"
  llm-mock run --mock-mcp slow_01=1s-10s --mock-mcp slow_02=hang codex

  -h, --help
        Show this help
`

// PrependRunFlagsFromEnv reads LLM_MOCK_RUN_FLAGS and prepends tokenized flags to args.
func PrependRunFlagsFromEnv(args []string) []string {
	env := os.Getenv(RunFlagsEnvVar)
	if env == "" {
		return args
	}
	flags := strings.Fields(env)
	if len(flags) == 0 {
		return args
	}
	out := make([]string, 0, len(flags)+len(args))
	out = append(out, flags...)
	out = append(out, args...)
	return out
}

// ParseRunFlagsFromEnv parses llm-mock run flags from LLM_MOCK_RUN_FLAGS only.
// Shortcut binaries (llm-mock-run-grok) use this so all argv is forwarded to the agent.
func ParseRunFlagsFromEnv() (RunGrokOptions, error) {
	env := strings.TrimSpace(os.Getenv(RunFlagsEnvVar))
	if env == "" {
		return RunGrokOptions{}, nil
	}
	opts, _, err := parseRunFlags(strings.Fields(env), false)
	return opts, err
}

// ParseRunFlags parses --mock-events-preset, --log-events, and --log-http from args
// (lessflags + StopOnFirstArg). Returns RunGrokOptions, remaining args, and error.
func ParseRunFlags(args []string) (RunGrokOptions, []string, error) {
	return parseRunFlags(args, true)
}

func parseRunFlags(args []string, stopOnFirstArg bool) (RunGrokOptions, []string, error) {
	var logEvents, logHTTP, mockEventsPreset, mockEventsFile, agentRunnerConfigHome *string
	var mockMCP []string
	builder := lessflags.String("--mock-events-preset", &mockEventsPreset).
		String("--mock-events-file", &mockEventsFile).
		StringSlice("--mock-mcp", &mockMCP).
		String("--log-events", &logEvents).
		String("--log-http", &logHTTP).
		String("--agent-runner-config-home", &agentRunnerConfigHome).
		Help("-h,--help", RunCommandHelp).
		HelpNoExit()
	if stopOnFirstArg {
		builder = builder.StopOnFirstArg()
	}
	remain, err := builder.Parse(args)
	if err != nil {
		return RunGrokOptions{}, nil, err
	}
	opts := RunGrokOptions{}
	if mockEventsPreset != nil {
		opts.MockEventsPreset = *mockEventsPreset
	}
	if mockEventsFile != nil {
		opts.MockEventsFile = *mockEventsFile
	}
	if logEvents != nil {
		opts.LogEventsPath = *logEvents
	}
	if logHTTP != nil {
		opts.LogHTTPPath = *logHTTP
	}
	if agentRunnerConfigHome != nil {
		opts.AgentRunnerConfigHome = *agentRunnerConfigHome
	}
	opts.MockMCP = mockMCP
	return opts, remain, nil
}
