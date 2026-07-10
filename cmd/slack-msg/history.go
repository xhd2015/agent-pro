package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"

	lessflags "github.com/xhd2015/less-flags"
	"github.com/slack-go/slack"
	"github.com/xhd2015/agent-pro/pkgs/slackutil"
)

const historyHelpText = `slack-msg history: fetch conversation history or thread replies.

Usage:
  slack-msg history [options] [CHANNEL]

Options:
  --token TOKEN     Bot token (env: SLACK_BOT_TOKEN)
  --channel CHANNEL Channel ID or name (env: SLACK_CHANNEL)
  --config PATH     JSON config file (env: SLACK_CONFIG)
  --limit N         Max messages to fetch
  --thread TS       Fetch thread replies for TS
  --json            Structured JSON output
  -h, --help        Show help
`

type historyMsg struct {
	TS       string `json:"ts"`
	User     string `json:"user"`
	Text     string `json:"text"`
	ThreadTS string `json:"thread_ts,omitempty"`
}

type historyDoc struct {
	Messages []historyMsg `json:"messages"`
	HasMore  bool         `json:"has_more"`
}

func runHistory(args []string) error {
	var (
		tokenFlag   *string
		channelFlag *string
		configFlag  *string
		limitFlag   *int
		threadFlag  *string
		jsonFlag    bool
	)

	remain, err := lessflags.String("--token", &tokenFlag).
		String("--channel", &channelFlag).
		String("--config", &configFlag).
		Int("--limit", &limitFlag).
		String("--thread", &threadFlag).
		Bool("--json", &jsonFlag).
		Help("-h,--help", historyHelpText).
		HelpNoExit().
		StopOnFirstArg().
		Parse(args)
	if errors.Is(err, lessflags.ErrHelp) {
		return nil
	}
	if err != nil {
		return err
	}

	configPath := ""
	if configFlag != nil && *configFlag != "" {
		configPath = *configFlag
	} else if env := os.Getenv("SLACK_CONFIG"); env != "" {
		configPath = env
	}

	var cfg *slackutil.SlackConfig
	if configPath != "" {
		loaded, loadErr := slackutil.Load(configPath)
		if loadErr != nil {
			fmt.Fprintf(os.Stderr, "failed to load config: %v\n", loadErr)
			os.Exit(1)
		}
		cfg = loaded
	}

	token := slackutil.ResolveBotToken(flagString(tokenFlag), "SLACK_BOT_TOKEN", cfg)
	if token == "" {
		if cfg != nil && configPath != "" {
			fmt.Fprintf(os.Stderr, "botToken is empty in %s\n", slackutil.ConfigDisplayPath(configPath))
		} else {
			fmt.Fprintln(os.Stderr, "bot token required")
		}
		os.Exit(1)
	}

	channel := ""
	if channelFlag != nil && *channelFlag != "" {
		channel = *channelFlag
	} else if len(remain) >= 1 {
		channel = remain[0]
	} else if env := os.Getenv("SLACK_CHANNEL"); env != "" {
		channel = env
	} else if cfg != nil {
		channel = cfg.DefaultChannelId
	}
	if channel == "" {
		fmt.Fprintln(os.Stderr, "channel required")
		os.Exit(1)
	}
	if len(remain) > 1 {
		return fmt.Errorf("unexpected arguments: %v", remain[1:])
	}

	api := slackutil.NewAPIClient(token)

	resolvedChannel, resolveErr := slackutil.ResolveChannel(api, cfg, channel)
	if resolveErr != nil {
		fmt.Fprintf(os.Stderr, "history failed: %v\n", resolveErr)
		os.Exit(1)
	}

	limit := 0
	if limitFlag != nil && *limitFlag > 0 {
		limit = *limitFlag
	}

	var (
		messages []slack.Message
		hasMore  bool
	)

	// Fetch a page from Slack, then sort oldest→newest and apply --limit client-side
	// as the N newest messages. Passing limit only to the API is insufficient when
	// mocks return unsorted pages; client-side ensures chronological newest-N.
	threadTS := flagString(threadFlag)
	if threadTS != "" {
		params := &slack.GetConversationRepliesParameters{
			ChannelID: resolvedChannel,
			Timestamp: threadTS,
		}
		msgs, more, _, repliesErr := api.GetConversationReplies(params)
		if repliesErr != nil {
			fmt.Fprintf(os.Stderr, "history failed: %v\n", repliesErr)
			os.Exit(1)
		}
		messages = msgs
		hasMore = more
	} else {
		params := &slack.GetConversationHistoryParameters{
			ChannelID: resolvedChannel,
		}
		resp, histErr := api.GetConversationHistory(params)
		if histErr != nil {
			fmt.Fprintf(os.Stderr, "history failed: %v\n", histErr)
			os.Exit(1)
		}
		messages = resp.Messages
		hasMore = resp.HasMore
	}

	// Chronological: oldest → newest (API returns newest-first / arbitrary mock order).
	sort.SliceStable(messages, func(i, j int) bool {
		return messages[i].Timestamp < messages[j].Timestamp
	})

	// --limit N: keep the N newest messages, still printed oldest→newest.
	if limit > 0 && len(messages) > limit {
		messages = messages[len(messages)-limit:]
	}

	out := make([]historyMsg, 0, len(messages))
	for _, m := range messages {
		out = append(out, historyMsg{
			TS:       m.Timestamp,
			User:     m.User,
			Text:     m.Text,
			ThreadTS: m.ThreadTimestamp,
		})
	}

	if jsonFlag {
		doc := historyDoc{Messages: out, HasMore: hasMore}
		enc := json.NewEncoder(os.Stdout)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(doc); err != nil {
			fmt.Fprintf(os.Stderr, "history failed: %v\n", err)
			os.Exit(1)
		}
		// json.Encoder.Encode already appends a trailing newline.
		return nil
	}

	for _, m := range out {
		fmt.Printf("[%s] %s: %s\n", m.TS, m.User, m.Text)
	}
	// When there are zero messages, still emit a trailing newline for CLI consistency.
	if len(out) == 0 {
		fmt.Println()
	}
	return nil
}
