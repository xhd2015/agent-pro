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
	b.WriteString("Read thread history:\n")
	b.WriteString("```\n")
	b.WriteString(fmt.Sprintf("slack-msg history --channel %s --thread %s\n", ctx.ChannelID, ctx.ThreadTS))
	b.WriteString("```\n\n")
	b.WriteString("Reply on Slack (same channel + thread):\n")
	b.WriteString("```\n")
	b.WriteString(fmt.Sprintf("slack-msg send --channel %s --thread %s \"your answer\"\n", ctx.ChannelID, ctx.ThreadTS))
	b.WriteString("```\n\n")
	b.WriteString("## Workflow\n\n")
	b.WriteString("1. Investigate the request\n")
	b.WriteString("2. Reply via slack-msg send (thread_ts required for channel threads)\n\n")
	b.WriteString("## Notes\n\n")
	b.WriteString("Use host env/config for credentials. Do not print secrets.\n")
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
