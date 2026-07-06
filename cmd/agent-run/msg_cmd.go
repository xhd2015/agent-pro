package main

import (
	"fmt"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/agentsend"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/less-gen/flags"
)

const msgHelp = `
Usage: agent-run msg <subcommand> <session-id>/<message-id>

Manage queued send messages for a live grok-tty or codex-tty session.

Subcommands:
  status   print message status (pending or delivered)
  cancel   remove a pending message from the send queue

Options:
  -h, --help   show help
`

const msgStatusHelp = `
Usage: agent-run msg status <session-id>/<message-id>

Print message status to stdout:
  pending    message is still in the send queue
  delivered  message is no longer queued (delivered, cancelled, or timed out)

Options:
  -h, --help   show help
`

const msgCancelHelp = `
Usage: agent-run msg cancel <session-id>/<message-id>

Remove a pending message from the send queue.

Options:
  -h, --help   show help
`

func runMsg(args []string) error {
	if len(args) == 0 {
		fmt.Print(strings.TrimPrefix(msgHelp, "\n"))
		return nil
	}
	switch args[0] {
	case "status":
		return runMsgStatus(args[1:])
	case "cancel":
		return runMsgCancel(args[1:])
	case "-h", "--help":
		fmt.Print(strings.TrimPrefix(msgHelp, "\n"))
		return nil
	default:
		return fmt.Errorf("unknown msg subcommand: %s", args[0])
	}
}

func runMsgStatus(args []string) error {
	remaining, err := flags.Help("-h,--help", msgStatusHelp).Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) != 1 {
		return fmt.Errorf("msg status: requires <session-id>/<message-id>")
	}
	sess, msgID, err := resolveMsgSession(remaining[0])
	if err != nil {
		return err
	}
	status, err := agentsend.MessageStatus(sess.Home, sess, msgID)
	if err != nil {
		return err
	}
	fmt.Println(status)
	return nil
}

func runMsgCancel(args []string) error {
	remaining, err := flags.Help("-h,--help", msgCancelHelp).Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) != 1 {
		return fmt.Errorf("msg cancel: requires <session-id>/<message-id>")
	}
	sess, msgID, err := resolveMsgSession(remaining[0])
	if err != nil {
		return err
	}
	return agentsend.Cancel(sess.Home, sess, msgID)
}

func parseSessionMessageRef(arg string) (sessionID, msgID string, err error) {
	parts := strings.SplitN(arg, "/", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return "", "", fmt.Errorf("requires <session-id>/<message-id>")
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
}

func resolveMsgSession(ref string) (agentsend.Session, string, error) {
	sessionID, msgID, err := parseSessionMessageRef(ref)
	if err != nil {
		return agentsend.Session{}, "", err
	}

	store, err := openStore()
	if err != nil {
		return agentsend.Session{}, "", err
	}
	home := store.Home()

	ttySess, err := agenttty.ResolveByTerminalID(home, sessionID)
	if err != nil {
		return agentsend.Session{}, "", err
	}

	return agentsend.Session{
		Home:              home,
		Runner:            ttySess.RunnerID,
		TerminalSessionID: sessionID,
		ListenAddr:        ttySess.Registry.ListenAddr,
	}, msgID, nil
}