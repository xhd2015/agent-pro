package main

import (
	"fmt"

	"github.com/xhd2015/agent-pro/pkgs/groktty"
	ptyclient "github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap/client"
	"github.com/xhd2015/less-gen/flags"
)

const attachHelp = `
Usage: agent-run attach <session-id>

Attach to a live grok-tty or codex-tty session by registry id (printed on stderr during run).

Options:
  -h, --help          show help
`

func runAttach(args []string) error {
	remaining, err := flags.Help("-h,--help", attachHelp).Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) != 1 {
		return fmt.Errorf("attach: requires <session-id>")
	}
	sessionID := remaining[0]

	store, err := openStore()
	if err != nil {
		return err
	}
	entry, _, err := groktty.ReadSupportedRegistry(store.Home(), sessionID)
	if err != nil {
		return err
	}

	c := ptyclient.NewClient("http://" + entry.ListenAddr)
	_, err = ptyclient.Attach(c, ptyclient.ConnectOptions{
		SessionID:      sessionID,
		AttachSnapshot: true,
		Wait:           true,
	})
	return err
}
