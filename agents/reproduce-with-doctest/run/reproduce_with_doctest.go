// Package run implements the reproduce-with-doctest sub-agent CLI.
//
// The reproduce-with-doctest agent encodes bugs as failing doctest cases in
// existing test trees. It
// auto-detects the parent agent runner (opencode, pi, codex, crush, grok)
// and delegates execution via the subagent package.
//
// Usage:
//
//	reproduce-with-doctest [OPTIONS] <prompt>
//
// Options:
//
//	--agent-runner <id>      override agent runner (opencode|pi|codex|crush|grok)
//	--model <model>          override model
//	--model-env <env>        override the env var used to pass the model
//	--session-id <id>        resume an existing session
//	--timeout <duration>     timeout (default: 1h, min: 1m)
//	--catch-up               replay session events
//	--status                 show session status
//	--list-sessions          list all sessions
//	--session-base <dir>     override sessions directory
//	-h, --help               show help
package run

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/agent/subagent"
	"github.com/xhd2015/less-gen/flags"
	"github.com/xhd2015/skills/skill_file"
)

//go:embed SKILL.md
var SkillFile string

// getPrompt returns the embedded SKILL.md content with the YAML
// frontmatter header stripped, leaving only the system prompt body.
func getPrompt() string {
	return skill_file.TrimHeader(SkillFile)
}

var help = `
Usage: reproduce-with-doctest [OPTIONS] <prompt>

A doctest-backed bug reproduction specialist. Reproduces bugs by adding failing
doctest cases to existing test trees (SETUP.md + ASSERT.md). Requires at least
one RED doctest run proving expected vs actual behavior mismatch. Delegates to
the parent agent runner (auto-detected) for LLM calls.

Options:
  --agent-runner <id>      override agent runner (opencode|pi|codex|crush|grok)
  --model <model>          override model (e.g. "anthropic/claude-sonnet-4-20250514")
  --model-env <env>        override the env var used to pass the model
  --session-id <id>        resume an existing session
  --timeout <duration>     timeout (default: 1h, min: 1m, e.g. "30m", "2h")
  --catch-up               replay session events (requires --session-id)
  --status                 show session status (requires --session-id)
  --list-sessions          list all reproduce-with-doctest sessions
  --session-base <dir>     override sessions directory
  -h, --help               show this help
`

const defaultModelEnv = "AGENT_PRO_SUBAGENT_REPRODUCE_WITH_DOCTEST_MODEL"

// Config holds all customizable configuration for the reproduce-with-doctest agent.
// Zero-value fields fall back to sensible defaults.
type Config struct {
	ModelEnv     string        // env var name for model; empty = default
	AgentRunner  string        // override agent runner (opencode|pi|codex|crush|grok)
	Model        string        // model name (set via env var)
	SessionID    string        // resume an existing session
	Timeout      time.Duration // 0 = default 1h
	CatchUp      bool          // replay session events
	Status       bool          // show session status
	ListSessions bool          // list all sessions
	SessionBase  string        // override sessions directory
	Prompt       string        // user prompt
}

// Run runs the reproduce-with-doctest agent programmatically with the given config.
func Run(ctx context.Context, cfg Config) error {
	modelEnv := cfg.ModelEnv
	if modelEnv == "" {
		modelEnv = defaultModelEnv
	}

	// Model is passed via env var so the subagent runner picks it up.
	if cfg.Model != "" {
		os.Setenv(modelEnv, cfg.Model)
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = time.Hour
	}

	c := subagent.Config{
		RoleName:      "reproduce-with-doctest",
		Cmd:           "reproduce-with-doctest",
		PromptContent: getPrompt(),
		ModelEnv:      modelEnv,
	}

	opts := subagent.Options{
		Prompt:       cfg.Prompt,
		AgentRunner:  cfg.AgentRunner,
		SessionID:    cfg.SessionID,
		CatchUp:      cfg.CatchUp,
		Status:       cfg.Status,
		ListSessions: cfg.ListSessions,
		SessionBase:  cfg.SessionBase,
		Timeout:      timeout,
	}

	return subagent.Run(ctx, c, opts)
}

// RunArgs parses CLI arguments into a Config and delegates to Run.
func RunArgs(args []string) error {
	var agentRunner *string
	var model *string
	var modelEnv *string
	var sessionID *string
	var timeoutStr *string
	var catchUp *bool
	var status *bool
	var listSessions *bool
	var sessionBase *string

	remaining, err := flags.
		String("--agent-runner", &agentRunner).
		String("--model", &model).
		String("--model-env", &modelEnv).
		String("--session-id", &sessionID).
		String("--timeout", &timeoutStr).
		Bool("--catch-up", &catchUp).
		Bool("--status", &status).
		Bool("--list-sessions", &listSessions).
		String("--session-base", &sessionBase).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	cfg := Config{
		Prompt: strings.Join(remaining, " "),
	}

	if agentRunner != nil {
		cfg.AgentRunner = strings.TrimSpace(*agentRunner)
	}
	if model != nil {
		cfg.Model = strings.TrimSpace(*model)
	}
	if modelEnv != nil {
		cfg.ModelEnv = strings.TrimSpace(*modelEnv)
	}
	if sessionID != nil {
		cfg.SessionID = strings.TrimSpace(*sessionID)
	}
	if timeoutStr != nil {
		d, err := subagent.ParseTimeoutDuration(strings.TrimSpace(*timeoutStr))
		if err != nil {
			return fmt.Errorf("invalid --timeout: %w", err)
		}
		cfg.Timeout = d
	}
	if catchUp != nil && *catchUp {
		cfg.CatchUp = true
	}
	if status != nil && *status {
		cfg.Status = true
	}
	if listSessions != nil && *listSessions {
		cfg.ListSessions = true
	}
	if sessionBase != nil {
		cfg.SessionBase = strings.TrimSpace(*sessionBase)
	}

	return Run(context.Background(), cfg)
}