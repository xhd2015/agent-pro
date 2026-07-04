package main

import (
	"context"
	"fmt"

	"github.com/xhd2015/agent-pro/pkgs/ttywatch"
)

func runServeSession(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("serve: missing session id or command")
	}
	sessionID := args[0]
	command := args[1:]
	return ttywatch.ServeSession(context.Background(), ttywatch.ServeOptions{
		SessionID: sessionID,
		Command:   command,
	})
}