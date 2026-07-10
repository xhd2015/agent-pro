package main

import (
	"errors"
	"fmt"
	"os"

	lessflags "github.com/xhd2015/less-flags"
	"github.com/slack-go/slack"
	"github.com/xhd2015/agent-pro/pkgs/slackutil"
)

const sendHelpText = `slack-msg send: post a message via Slack Web API.

Usage:
  slack-msg send [options] MESSAGE

Options:
  --token TOKEN     Bot token (env: SLACK_BOT_TOKEN)
  --channel CHANNEL Channel ID or name (env: SLACK_CHANNEL)
  --config PATH     JSON config file (env: SLACK_CONFIG)
  --thread TS       Optional thread timestamp
  -h, --help        Show help
`

func runSend(args []string) error {
	var tokenFlag, channelFlag, configFlag, threadFlag *string

	remain, err := lessflags.String("--token", &tokenFlag).
		String("--channel", &channelFlag).
		String("--config", &configFlag).
		String("--thread", &threadFlag).
		Help("-h,--help", sendHelpText).
		HelpNoExit().
		StopOnFirstArg().
		Parse(args)
	if errors.Is(err, lessflags.ErrHelp) {
		return nil
	}
	if err != nil {
		return err
	}

	switch len(remain) {
	case 0:
		fmt.Fprintln(os.Stderr, "message required")
		os.Exit(1)
	case 1:
		// ok
	default:
		fmt.Fprintf(os.Stderr, "exactly one message required, got %d\n", len(remain))
		os.Exit(1)
	}
	message := remain[0]

	configPath := ""
	if configFlag != nil && *configFlag != "" {
		configPath = *configFlag
	} else if env := os.Getenv("SLACK_CONFIG"); env != "" {
		configPath = env
	}

	configDisplay := "(none)"
	var cfg *slackutil.SlackConfig
	if configPath != "" {
		configDisplay = slackutil.ConfigDisplayPath(configPath)
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
			fmt.Fprintf(os.Stderr, "botToken is empty in %s\n", configDisplay)
		} else {
			fmt.Fprintln(os.Stderr, "bot token required")
		}
		os.Exit(1)
	}

	channel := ""
	if channelFlag != nil && *channelFlag != "" {
		channel = *channelFlag
	} else if env := os.Getenv("SLACK_CHANNEL"); env != "" {
		channel = env
	} else if cfg != nil {
		channel = cfg.DefaultChannelId
	}
	if channel == "" {
		fmt.Fprintln(os.Stderr, "channel required")
		os.Exit(1)
	}

	api := slackutil.NewAPIClient(token)

	resolvedChannel, resolveErr := slackutil.ResolveChannel(api, cfg, channel)
	if resolveErr != nil {
		fmt.Fprintf(os.Stderr, "send failed: %v\n", resolveErr)
		os.Exit(1)
	}

	fmt.Printf("Sending to channel=%s: %q\n", resolvedChannel, message)
	fmt.Printf("Using config from: %s\n", configDisplay)

	opts := []slack.MsgOption{slack.MsgOptionText(message, false)}
	if thread := flagString(threadFlag); thread != "" {
		opts = append(opts, slack.MsgOptionTS(thread))
	}

	respChannel, ts, err := api.PostMessage(resolvedChannel, opts...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "send failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("OK ts=%s channel=%s\n", ts, respChannel)
	return nil
}
