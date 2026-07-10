package slackutil

import (
	"fmt"
	"strings"

	"github.com/slack-go/slack"
)

// IsChannelID reports whether channel looks like a Slack channel/DM/group ID.
func IsChannelID(channel string) bool {
	return strings.HasPrefix(channel, "C") || strings.HasPrefix(channel, "D") || strings.HasPrefix(channel, "G")
}

// LookupKnownChannel resolves a channel name via the knownChannels map.
func LookupKnownChannel(known map[string]string, name string) (string, bool) {
	if known == nil {
		return "", false
	}
	candidates := []string{name}
	if strings.HasPrefix(name, "#") {
		candidates = append(candidates, name[1:])
	} else {
		candidates = append(candidates, "#"+name)
	}
	for _, candidate := range candidates {
		if id, ok := known[candidate]; ok && id != "" {
			return id, true
		}
	}
	return "", false
}

func channelNameForLookup(name string) string {
	return strings.TrimPrefix(name, "#")
}

// ResolveChannel resolves a channel name or ID to a Slack channel ID.
func ResolveChannel(api *slack.Client, cfg *SlackConfig, channel string) (string, error) {
	if IsChannelID(channel) {
		return channel, nil
	}

	if cfg != nil {
		if id, ok := LookupKnownChannel(cfg.KnownChannels, channel); ok {
			return id, nil
		}
	}

	name := channelNameForLookup(channel)
	params := &slack.GetConversationsParameters{
		Types: []string{"public_channel", "private_channel"},
		Limit: 200,
	}
	for {
		channels, nextCursor, err := api.GetConversations(params)
		if err != nil {
			return "", fmt.Errorf("channel lookup failed: %w", err)
		}
		for _, ch := range channels {
			if ch.Name == name {
				return ch.ID, nil
			}
		}
		if nextCursor == "" {
			break
		}
		params.Cursor = nextCursor
	}

	return "", fmt.Errorf("channel not found: %s", channel)
}