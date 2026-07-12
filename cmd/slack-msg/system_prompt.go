package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// defaultSlackLocalBotRelDir is the relative path under $HOME for session playbooks.
const defaultSlackLocalBotRelDir = ".agent-pro/slack-local-bot"

// defaultListenLockRelPath is the relative path under $HOME for the singleton lock.
const defaultListenLockRelPath = ".agent-pro/slack-msg.listen.lock"

func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		home = os.Getenv("HOME")
	}
	return home
}

func defaultListenLockPath() string {
	return filepath.Join(homeDir(), filepath.FromSlash(defaultListenLockRelPath))
}

func defaultSlackLocalBotRoot() string {
	return filepath.Join(homeDir(), filepath.FromSlash(defaultSlackLocalBotRelDir))
}

func sessionSystemMDPath(dataRoot, sessionID string) string {
	return filepath.Join(dataRoot, "sessions", sessionID, "SYSTEM.md")
}

type systemPromptContext struct {
	SessionID string
	ChannelID string
	ThreadTS  string
}

type openInjectContext struct {
	SessionID        string
	ChannelID        string
	ThreadTS         string
	FromDisplay      string
	FromUserID       string
	SystemPromptPath string
	UserMessage      string
}

// formatSystemPrompt builds SYSTEM.md playbook body (no secrets).
// Recipes use session-bound CLI only (no raw send --channel/--thread).
func formatSystemPrompt(ctx systemPromptContext) string {
	var b strings.Builder
	b.WriteString("# Slack local-bot agent session\n\n")
	b.WriteString("You are the assistant for a Slack listen session.\n\n")
	b.WriteString("## Mission\n\n")
	b.WriteString("Investigate the master's request. Reply with an answer or a clarification using the CLI below.\n\n")
	b.WriteString("## This session\n\n")
	b.WriteString("- session-id: ")
	b.WriteString(ctx.SessionID)
	b.WriteString("\n")
	b.WriteString("- channel: ")
	b.WriteString(ctx.ChannelID)
	b.WriteString("\n")
	b.WriteString("- thread_ts: ")
	b.WriteString(ctx.ThreadTS)
	b.WriteString("\n\n")
	b.WriteString("## CLI recipes\n\n")
	b.WriteString("Read local session history:\n")
	b.WriteString("```\n")
	b.WriteString("slack-msg session history\n")
	b.WriteString("```\n\n")
	b.WriteString("Read history after a message id:\n")
	b.WriteString("```\n")
	b.WriteString("slack-msg session history --after-msg-id <MSG_ID>\n")
	b.WriteString("```\n\n")
	b.WriteString("Reply on Slack (channel top-level):\n")
	b.WriteString("```\n")
	b.WriteString("slack-msg session reply \"your answer\"\n")
	b.WriteString("```\n\n")
	b.WriteString("## Workflow\n\n")
	b.WriteString("1. Investigate the request\n")
	b.WriteString("2. Optionally read session history\n")
	b.WriteString("3. Reply via slack-msg session reply\n\n")
	b.WriteString("## Notes\n\n")
	b.WriteString("Session id and config are injected via env (SLACK_MSG_SESSION_ID, SLACK_MSG_CONFIG).\n")
	b.WriteString("Use only the session recipes above. Do not print secrets.\n")
	return b.String()
}

// writeSystemPrompt writes SYSTEM.md, creating parent directories as needed.
func writeSystemPrompt(path string, ctx systemPromptContext) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("system prompt path is empty")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(formatSystemPrompt(ctx)), 0o644)
}

// formatOpenInject builds the short OpenSession inject prompt for thread mode.
func formatOpenInject(ctx openInjectContext) string {
	from := strings.TrimSpace(ctx.FromDisplay)
	if from == "" {
		from = ctx.FromUserID
	}
	if ctx.FromUserID != "" && from != ctx.FromUserID {
		from = fmt.Sprintf("%s (%s)", from, ctx.FromUserID)
	} else if ctx.FromUserID != "" {
		from = ctx.FromUserID
	}

	var b strings.Builder
	b.WriteString("Slack listen session open\n")
	b.WriteString("session-id: ")
	b.WriteString(ctx.SessionID)
	b.WriteString("\n")
	b.WriteString("channel: ")
	b.WriteString(ctx.ChannelID)
	b.WriteString("\n")
	b.WriteString("thread_ts: ")
	b.WriteString(ctx.ThreadTS)
	b.WriteString("\n")
	b.WriteString("from: ")
	b.WriteString(from)
	b.WriteString("\n")
	b.WriteString("Instructions: ")
	b.WriteString(ctx.SystemPromptPath)
	b.WriteString("\n")
	b.WriteString("User message:\n")
	b.WriteString(ctx.UserMessage)
	if !strings.HasSuffix(ctx.UserMessage, "\n") {
		b.WriteString("\n")
	}
	return b.String()
}

// stripBotMention removes the listening bot's <@USERID> mention and collapses whitespace.
func stripBotMention(text, botUserID string) string {
	if botUserID != "" {
		text = strings.ReplaceAll(text, "<@"+botUserID+">", "")
	}
	return strings.Join(strings.Fields(text), " ")
}
