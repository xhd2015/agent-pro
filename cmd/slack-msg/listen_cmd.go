package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	lessflags "github.com/xhd2015/less-flags"
	"github.com/xhd2015/agent-pro/pkgs/slackutil"
)

const listenHelpText = `slack-msg listen: Slack Socket Mode inbound bridge.

Usage:
  slack-msg listen [options]

Options:
  --token TOKEN
  --app-token TOKEN
  --config PATH
  --channel CHANNEL
  --require-mention
  --no-require-mention
  --allow-from USER_ID
  --session-mode MODE
  --idle-timeout DURATION
  --agent-runner RUNNER
  --agent-runner-config-home PATH
  --reply-prefix TEXT
  --lock-file PATH
  -h, --help
`

func runListenCommand(args []string) error {
	var (
		tokenFlag                 *string
		appTokenFlag              *string
		configFlag                *string
		channels                  []string
		requireMention            = true
		noRequireMention          bool
		requireMentionFlag        bool
		allowFrom                 []string
		sessionModeFlag           *string
		idleTimeoutFlag           *string
		agentRunnerFlag           *string
		agentRunnerConfigHomeFlag *string
		replyPrefixFlag           *string
		lockFileFlag              *string
	)

	remain, err := lessflags.String("--token", &tokenFlag).
		String("--app-token", &appTokenFlag).
		String("--config", &configFlag).
		StringSlice("--channel", &channels).
		Bool("--require-mention", &requireMentionFlag).
		Bool("--no-require-mention", &noRequireMention).
		StringSlice("--allow-from", &allowFrom).
		String("--session-mode", &sessionModeFlag).
		String("--idle-timeout", &idleTimeoutFlag).
		String("--agent-runner", &agentRunnerFlag).
		String("--agent-runner-config-home", &agentRunnerConfigHomeFlag).
		String("--reply-prefix", &replyPrefixFlag).
		String("--lock-file", &lockFileFlag).
		Help("-h,--help", listenHelpText).
		HelpNoExit().
		Parse(args)
	if errors.Is(err, lessflags.ErrHelp) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(remain) > 0 {
		return fmt.Errorf("unexpected arguments: %v", remain)
	}
	if noRequireMention {
		requireMention = false
	}
	if requireMentionFlag {
		requireMention = true
	}

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

	botToken := slackutil.ResolveBotToken(flagString(tokenFlag), "SLACK_BOT_TOKEN", cfg)
	if botToken == "" {
		fmt.Fprintln(os.Stderr, "bot token required")
		os.Exit(1)
	}

	appToken := slackutil.ResolveAppToken(flagString(appTokenFlag), "SLACK_APP_TOKEN", cfg)
	if appToken == "" {
		fmt.Fprintln(os.Stderr, "app token required")
		os.Exit(1)
	}

	sessionMode := "thread"
	if sessionModeFlag != nil && *sessionModeFlag != "" {
		sessionMode = *sessionModeFlag
	}

	idleTimeout := 30 * time.Minute
	if idleTimeoutFlag != nil && *idleTimeoutFlag != "" {
		parsed, parseErr := time.ParseDuration(*idleTimeoutFlag)
		if parseErr != nil {
			return fmt.Errorf("invalid --idle-timeout: %w", parseErr)
		}
		idleTimeout = parsed
	}

	agentRunner := "grok-tty"
	if agentRunnerFlag != nil && *agentRunnerFlag != "" {
		agentRunner = *agentRunnerFlag
	}

	return runListen(listenConfig{
		BotToken:              botToken,
		AppToken:              appToken,
		ConfigPath:            configPath,
		ConfigDisplay:         configDisplay,
		SlackConfig:           cfg,
		Channels:              channels,
		RequireMention:        requireMention,
		AllowFrom:             allowFrom,
		SessionMode:           sessionMode,
		IdleTimeout:           idleTimeout,
		AgentRunner:           agentRunner,
		AgentRunnerConfigHome: flagString(agentRunnerConfigHomeFlag),
		ReplyPrefix:           flagString(replyPrefixFlag),
		LockFile:              flagString(lockFileFlag),
	})
}
