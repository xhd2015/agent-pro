// Package main implements the reproduce sub-agent CLI.
//
// The reproduce agent is a bug reproduction specialist that excels at
// understanding reported issues, reproducing them locally, breaking down
// complex issues into discrete steps, and building minimal reproducible
// examples. It auto-detects the parent agent runner (opencode, pi, codex,
// crush) and delegates execution via the subagent package.
//
// Usage:
//
//	reproduce [OPTIONS] <prompt>
//
// Options:
//
//	--agent-runner <id>      override agent runner (opencode|pi|codex|crush)
//	--model <model>          override model
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
var skillFile string

// getPrompt returns the embedded SKILL.md content with the YAML
// frontmatter header stripped, leaving only the system prompt body.
func getPrompt() string {
	return skill_file.TrimHeader(skillFile)
}

var help = `
Usage: reproduce [OPTIONS] <prompt>

A bug reproduction specialist. Understands reported issues, attempts local
reproduction, breaks down complex issues into steps, and builds minimal
reproducible examples. Delegates to the parent agent runner (auto-detected)
for LLM calls.

Options:
  --agent-runner <id>      override agent runner (opencode|pi|codex|crush)
  --model <model>          override model (e.g. "anthropic/claude-sonnet-4-20250514")
  --session-id <id>        resume an existing session
  --timeout <duration>     timeout (default: 1h, min: 1m, e.g. "30m", "2h")
  --catch-up               replay session events (requires --session-id)
  --status                 show session status (requires --session-id)
  --list-sessions          list all reproduce sessions
  --session-base <dir>     override sessions directory
  -h, --help               show this help
`

func Run(args []string) error {
	var agentRunner *string
	var model *string
	var sessionID *string
	var timeoutStr *string
	var catchUp *bool
	var status *bool
	var listSessions *bool
	var sessionBase *string

	remaining, err := flags.
		String("--agent-runner", &agentRunner).
		String("--model", &model).
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

	cfg := subagent.Config{
		RoleName:      "reproduce",
		Cmd:           "reproduce",
		PromptContent: getPrompt(),
		ModelEnv:      "AGENT_PRO_SUBAGENT_REPRODUCE_MODEL",
	}

	var opts subagent.Options

	if agentRunner != nil {
		opts.AgentRunner = strings.TrimSpace(*agentRunner)
	}
	// Model is passed via env var so the subagent runner picks it up.
	if model != nil && strings.TrimSpace(*model) != "" {
		os.Setenv("AGENT_PRO_SUBAGENT_REPRODUCE_MODEL", strings.TrimSpace(*model))
	}
	if sessionID != nil {
		opts.SessionID = strings.TrimSpace(*sessionID)
	}
	if timeoutStr != nil {
		d, err := subagent.ParseTimeoutDuration(strings.TrimSpace(*timeoutStr))
		if err != nil {
			return fmt.Errorf("invalid --timeout: %w", err)
		}
		opts.Timeout = d
	} else {
		opts.Timeout = time.Hour
	}
	if catchUp != nil && *catchUp {
		opts.CatchUp = true
	}
	if status != nil && *status {
		opts.Status = true
	}
	if listSessions != nil && *listSessions {
		opts.ListSessions = true
	}
	if sessionBase != nil {
		opts.SessionBase = strings.TrimSpace(*sessionBase)
	}

	opts.Prompt = strings.Join(remaining, " ")

	ctx := context.Background()
	return subagent.Run(ctx, cfg, opts)
}
