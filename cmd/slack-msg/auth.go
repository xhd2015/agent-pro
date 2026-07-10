package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	lessflags "github.com/xhd2015/less-flags"
	"github.com/slack-go/slack"
	"github.com/xhd2015/agent-pro/pkgs/slackutil"
)

const authHelpText = `slack-msg auth: inspect bot or app token status.

Usage:
  slack-msg auth <command> [options]

Commands:
  status  Show bot or app token status

Options:
  -h, --help  Show help
`

const authStatusHelpText = `slack-msg auth status: show bot or app token status.

Usage:
  slack-msg auth status [options]
  slack-msg auth status --app [options]

Options:
  --token TOKEN       Bot token (env: SLACK_BOT_TOKEN)
  --app-token TOKEN   App-level token (env: SLACK_APP_TOKEN)
  --config PATH       JSON config file (env: SLACK_CONFIG)
  --json              Structured JSON output
  --app               Validate app-level token (Socket Mode / connections)
  -h, --help          Show help
`

const appAuthNote = "app-level token (Socket Mode / connections); not used for channels/send/history"

type botAuthStatusDoc struct {
	Config      string `json:"config"`
	Kind        string `json:"kind"`
	Ok          bool   `json:"ok"`
	Team        string `json:"team"`
	TeamID      string `json:"team_id"`
	User        string `json:"user"`
	UserID      string `json:"user_id"`
	BotID       string `json:"bot_id"`
	URL         string `json:"url"`
	TokenMasked string `json:"token_masked"`
}

type appAuthStatusDoc struct {
	Config      string `json:"config"`
	Kind        string `json:"kind"`
	Ok          bool   `json:"ok"`
	TokenMasked string `json:"token_masked"`
	Note        string `json:"note"`
}

func runAuth(args []string) error {
	if len(args) == 0 {
		fmt.Print(authHelpText)
		return nil
	}

	switch args[0] {
	case "-h", "--help":
		fmt.Print(authHelpText)
		return nil
	case "status":
		return runAuthStatus(args[1:])
	default:
		if strings.HasPrefix(args[0], "-") {
			return fmt.Errorf("unknown option: %s", args[0])
		}
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func runAuthStatus(args []string) error {
	var (
		tokenFlag    *string
		appTokenFlag *string
		configFlag   *string
		jsonFlag     bool
		appFlag      bool
	)

	remain, err := lessflags.String("--token", &tokenFlag).
		String("--app-token", &appTokenFlag).
		String("--config", &configFlag).
		Bool("--json", &jsonFlag).
		Bool("--app", &appFlag).
		Help("-h,--help", authStatusHelpText).
		HelpNoExit().
		StopOnFirstArg().
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

	configPath := ""
	if configFlag != nil && *configFlag != "" {
		configPath = *configFlag
	} else if env := os.Getenv("SLACK_CONFIG"); env != "" {
		configPath = env
	}

	configDisplay := slackutil.ConfigDisplayPath(configPath)
	var cfg *slackutil.SlackConfig
	if configPath != "" {
		loaded, loadErr := slackutil.Load(configPath)
		if loadErr != nil {
			fmt.Fprintf(os.Stderr, "failed to load config: %v\n", loadErr)
			os.Exit(1)
		}
		cfg = loaded
	}

	if appFlag {
		return runAuthStatusApp(flagString(appTokenFlag), cfg, configDisplay, jsonFlag)
	}
	return runAuthStatusBot(flagString(tokenFlag), cfg, configDisplay, jsonFlag)
}

func runAuthStatusBot(tokenFlag string, cfg *slackutil.SlackConfig, configDisplay string, asJSON bool) error {
	token := slackutil.ResolveBotToken(tokenFlag, "SLACK_BOT_TOKEN", cfg)
	if token == "" {
		fmt.Fprintln(os.Stderr, "bot token required")
		os.Exit(1)
	}

	if !asJSON {
		fmt.Printf("Using config from: %s\n", configDisplay)
	}

	api := slackutil.NewAPIClient(token)
	auth, err := api.AuthTest()
	if err != nil {
		fmt.Fprintf(os.Stderr, "auth failed: %v\n", err)
		os.Exit(1)
	}

	masked := maskToken(token)
	if asJSON {
		doc := botAuthStatusDoc{
			Config:      configDisplay,
			Kind:        "bot",
			Ok:          true,
			Team:        auth.Team,
			TeamID:      auth.TeamID,
			User:        auth.User,
			UserID:      auth.UserID,
			BotID:       auth.BotID,
			URL:         auth.URL,
			TokenMasked: masked,
		}
		return encodeJSON(doc)
	}

	fmt.Printf("kind: bot\n")
	fmt.Printf("ok: true\n")
	fmt.Printf("team: %s (%s)\n", auth.Team, auth.TeamID)
	fmt.Printf("user: %s (%s)\n", auth.User, auth.UserID)
	fmt.Printf("bot_id: %s\n", auth.BotID)
	fmt.Printf("url: %s\n", auth.URL)
	fmt.Printf("token: %s\n", masked)
	return nil
}

func runAuthStatusApp(appTokenFlag string, cfg *slackutil.SlackConfig, configDisplay string, asJSON bool) error {
	token := slackutil.ResolveAppToken(appTokenFlag, "SLACK_APP_TOKEN", cfg)
	if token == "" {
		fmt.Fprintln(os.Stderr, "app token required")
		os.Exit(1)
	}

	if !asJSON {
		fmt.Printf("Using config from: %s\n", configDisplay)
	}

	api := slackutil.NewAPIClient("", slack.OptionAppLevelToken(token))
	_, _, err := api.StartSocketModeContext(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "auth failed: %v\n", err)
		os.Exit(1)
	}

	masked := maskToken(token)
	if asJSON {
		doc := appAuthStatusDoc{
			Config:      configDisplay,
			Kind:        "app",
			Ok:          true,
			TokenMasked: masked,
			Note:        appAuthNote,
		}
		return encodeJSON(doc)
	}

	fmt.Printf("kind: app\n")
	fmt.Printf("ok: true\n")
	fmt.Printf("token: %s\n", masked)
	fmt.Printf("note: %s\n", appAuthNote)
	return nil
}

// maskToken returns type-prefix + "..." + last 4 chars (never the full secret).
// Examples: xoxb-slacktest-token → xoxb-...oken; xapp-... → xapp-...oken.
func maskToken(token string) string {
	if token == "" {
		return ""
	}
	last4 := token
	if len(token) >= 4 {
		last4 = token[len(token)-4:]
	}
	prefix := tokenTypePrefix(token)
	return prefix + "..." + last4
}

func tokenTypePrefix(token string) string {
	for _, p := range []string{"xoxb-", "xapp-", "xoxp-", "xoxa-", "xoxs-", "xoxc-", "xoxe-"} {
		if strings.HasPrefix(token, p) {
			return p
		}
	}
	if i := strings.Index(token, "-"); i >= 0 {
		return token[:i+1]
	}
	if len(token) >= 4 {
		return token[:4]
	}
	return token
}

func encodeJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		fmt.Fprintf(os.Stderr, "auth failed: %v\n", err)
		os.Exit(1)
	}
	// json.Encoder.Encode already appends a trailing newline.
	return nil
}
