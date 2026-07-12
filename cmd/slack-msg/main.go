package main

import (
	"fmt"
	"os"
	"strings"
)

const rootHelpText = `slack-msg: Slack messaging CLI.

Usage:
  slack-msg <command> [options]
  slack-msg --help [--topic TOPIC]

Commands:
  send      Post a message via Slack Web API
  history   Fetch conversation history or thread replies
  listen    Socket Mode inbound bridge to agent-run
  channels  List or search workspace channels
  auth      Inspect bot or app token status
  session   Session-bound reply and history

Help topics:
  add-missing-scope  How to grant missing OAuth scopes (e.g. groups:read)

Options:
  -h, --help     Show help
  --topic TOPIC  With --help, show a help topic
`

const helpTopicAddMissingScope = `slack-msg help: add-missing-scope

You cannot add OAuth scopes to an existing bot token from this CLI.

To grant a missing scope (for example groups:read for private channels):

  1. Open https://api.slack.com/apps and select your app
  2. Go to OAuth & Permissions
  3. Under Bot Token Scopes, add the needed scope (e.g. groups:read)
  4. Reinstall the app to the workspace (scopes apply only after reinstall)
  5. Update botToken in your config file, or set SLACK_BOT_TOKEN, to the new token
  6. Retry the command

Note: the official slack CLI also cannot grant scopes without reinstalling the app.
`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		fmt.Print(rootHelpText)
		return nil
	}

	// Global help / --topic (before treating as a command).
	if args[0] == "-h" || args[0] == "--help" || args[0] == "--topic" {
		return runRootHelp(args)
	}

	switch args[0] {
	case "send":
		return runSend(args[1:])
	case "history":
		return runHistory(args[1:])
	case "listen":
		return runListenCommand(args[1:])
	case "channels":
		return runChannels(args[1:])
	case "auth":
		return runAuth(args[1:])
	case "session":
		return runSession(args[1:])
	default:
		if strings.HasPrefix(args[0], "-") {
			return fmt.Errorf("unknown option: %s", args[0])
		}
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

// runRootHelp handles -h/--help and --topic at the root level (either order).
func runRootHelp(args []string) error {
	hasHelp := false
	topic := ""
	topicSet := false

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-h", "--help":
			hasHelp = true
		case "--topic":
			if i+1 >= len(args) {
				return fmt.Errorf("--topic requires a value")
			}
			topic = args[i+1]
			topicSet = true
			i++
		default:
			if strings.HasPrefix(args[i], "-") {
				return fmt.Errorf("unknown option: %s", args[i])
			}
			return fmt.Errorf("unexpected argument: %s", args[i])
		}
	}

	if topicSet && !hasHelp {
		return fmt.Errorf("--topic requires --help")
	}
	if hasHelp && topicSet {
		return printHelpTopic(topic)
	}
	fmt.Print(rootHelpText)
	return nil
}

func printHelpTopic(topic string) error {
	switch topic {
	case "add-missing-scope":
		fmt.Print(helpTopicAddMissingScope)
		return nil
	default:
		return fmt.Errorf("unknown help topic: %s", topic)
	}
}

func flagString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
