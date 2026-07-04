package run

import (
	"os"
	"strings"

	lessflags "github.com/xhd2015/less-flags"
)

const RunFlagsEnvVar = "LLM_MOCK_RUN_FLAGS"

// RunCommandHelp documents llm-mock run flags and environment variables.
const RunCommandHelp = `
Usage: llm-mock run [--mock-events-preset NAME] [--log-events FILE] [--log-http FILE] grok [grok-args...]

Start a background llm-mock HTTP server, configure an isolated GROK_HOME pointing
at the mock, and run grok in the foreground. The mock stops when grok exits.

Options:
  --mock-events-preset NAME
        Named AgentEvent sequence to seed mock genQueue after config exchanges,
        or "list" to print the preset catalog and exit (no grok, no mock server).
  --log-events FILE
        Append standard AgentEvent JSONL for each served mock response.
        FILE must end with .jsonl. Passed to the mock as --agent-events-file.
  --log-http FILE
        Append full HTTP request/response exchange JSONL for each mock HTTP call.
        FILE must end with .jsonl. Passed to the mock as --log-http.

Environment (orchestrator):
  LLM_MOCK_CONFIG_FILE, LLM_MOCK_CONFIG   Mock exchange config JSON path
  LLM_MOCK_EVENTS_FILE                    Optional input exchanges JSONL (appended)
  LLM_MOCK_GROK_HOME                      Explicit grok home (default: temp dir)
  LLM_MOCK_RUN_FLAGS                      Space-separated default run flags (prepended before argv; CLI wins on duplicate)
  LLM_MOCK_RUN_GROK_COMMAND               Replace grok executable (tests/plumbing)
  LLM_MOCK_RUN_GROK_DEBUG=1               Verbose orchestrator stderr debug logs

Config is optional; with no config env vars, uses inline default {"exchanges": []}.

Examples:
  llm-mock run --mock-events-preset=list
  llm-mock run grok
  llm-mock run --mock-events-preset=think-message grok
  llm-mock run --log-events session.jsonl grok
  llm-mock run --log-http http.jsonl grok
  llm-mock run --log-events session.jsonl --log-http http.jsonl grok -p hello --always-approve

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

// ParseRunFlags parses --mock-events-preset, --log-events, and --log-http from args
// (lessflags + StopOnFirstArg). Returns RunGrokOptions, remaining args, and error.
func ParseRunFlags(args []string) (RunGrokOptions, []string, error) {
	var logEvents, logHTTP, mockEventsPreset *string
	remain, err := lessflags.String("--mock-events-preset", &mockEventsPreset).
		String("--log-events", &logEvents).
		String("--log-http", &logHTTP).
		Help("-h,--help", RunCommandHelp).
		StopOnFirstArg().
		Parse(args)
	if err != nil {
		return RunGrokOptions{}, nil, err
	}
	opts := RunGrokOptions{}
	if mockEventsPreset != nil {
		opts.MockEventsPreset = *mockEventsPreset
	}
	if logEvents != nil {
		opts.LogEventsPath = *logEvents
	}
	if logHTTP != nil {
		opts.LogHTTPPath = *logHTTP
	}
	return opts, remain, nil
}