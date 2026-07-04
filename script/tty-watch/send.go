package main

import (
	"fmt"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/ttyrunner"
)

func runSend(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("send: requires <session-id> and <message>")
	}
	sessionID := args[0]
	message := strings.Join(args[1:], " ")

	home, err := TTYWatchHome()
	if err != nil {
		return err
	}
	entry, err := ReadRegistry(home, sessionID)
	if err != nil {
		return err
	}
	if !tcpReachable(entry.ListenAddr) {
		RemoveRegistry(home, sessionID)
		return fmt.Errorf("tty-watch session %s not found", sessionID)
	}

	if err := prepareSessionInjectMode(entry.ListenAddr, sessionID); err != nil {
		return err
	}
	if err := ttyrunner.InjectInput(entry.ListenAddr, sessionID, []byte(message)); err != nil {
		return err
	}
	return nil
}