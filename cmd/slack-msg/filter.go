package main

import (
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/slackutil"
)

type inboundMessage struct {
	EventType string
	ChannelID string
	UserID    string
	Text      string
	TS        string
	ThreadTS  string
}

type filterConfig struct {
	BotUserID      string
	RequireMention bool
	AllowFrom      []string
	ChannelFilter  []string
}

func isDirectMessage(channelID string) bool {
	return slackutil.IsChannelID(channelID) && strings.HasPrefix(channelID, "D")
}

func allowFromPermits(allowFrom []string, userID string) bool {
	if len(allowFrom) == 0 {
		return true
	}
	for _, allowed := range allowFrom {
		if allowed == "*" {
			return true
		}
		if allowed == userID {
			return true
		}
	}
	return false
}

func channelFilterPermits(channelFilter []string, channelID string) bool {
	if len(channelFilter) == 0 {
		return true
	}
	for _, allowed := range channelFilter {
		if allowed == channelID {
			return true
		}
	}
	return false
}

func textMentionsBot(text, botUserID string) bool {
	if botUserID == "" {
		return false
	}
	return strings.Contains(text, "<@"+botUserID+">")
}

func shouldProcess(msg inboundMessage, cfg filterConfig) bool {
	if msg.UserID == "" {
		return false
	}
	if cfg.BotUserID != "" && msg.UserID == cfg.BotUserID {
		return false
	}
	if !allowFromPermits(cfg.AllowFrom, msg.UserID) {
		return false
	}
	if !channelFilterPermits(cfg.ChannelFilter, msg.ChannelID) {
		return false
	}

	switch msg.EventType {
	case "app_mention":
		return true
	case "message":
		if isDirectMessage(msg.ChannelID) {
			return true
		}
		// Thread replies in channels bypass requireMention (follow-ups use send).
		if msg.ThreadTS != "" {
			return true
		}
		if !cfg.RequireMention {
			return true
		}
		return textMentionsBot(msg.Text, cfg.BotUserID)
	default:
		return false
	}
}

func rootThreadTS(msg inboundMessage) string {
	if msg.ThreadTS != "" {
		return msg.ThreadTS
	}
	return msg.TS
}

func sessionID(channelID, threadTS string) string {
	return "slack-" + channelID + "-" + threadTS
}
