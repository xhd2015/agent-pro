package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/agentsend"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
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
		OnListening: func(ctx context.Context, listenAddr, home, registrySubdir string) {
			startServeSendQueueDrainer(ctx, sessionID, listenAddr, home, registrySubdir)
		},
	})
}

// startServeSendQueueDrainer runs the session-owned send-queue consumer in the
// detached __serve__ child so --no-wait enqueues deliver without CLI lifetime.
func startServeSendQueueDrainer(ctx context.Context, sessionID, listenAddr, serveHome, registrySubdir string) {
	home := strings.TrimSpace(os.Getenv("AGENT_RUN_HOME"))
	if home == "" {
		home = serveHome
	}
	if home == "" || sessionID == "" || listenAddr == "" {
		return
	}
	provider, ok := providerForRegistrySubdir(registrySubdir)
	if !ok {
		return
	}
	sess := agentsend.Session{
		Home:              home,
		Runner:            provider.ID,
		TerminalSessionID: sessionID,
		ListenAddr:        listenAddr,
	}
	agentsend.RunSessionDrainer(ctx, home, sess, provider)
}

func providerForRegistrySubdir(subdir string) (agenttty.Provider, bool) {
	subdir = strings.TrimSpace(subdir)
	if subdir == "" {
		return agenttty.Provider{}, false
	}
	for _, p := range agenttty.ProviderListSorted() {
		if p.RegistryDir == subdir {
			return p, true
		}
	}
	// Convention: "<runner-id>-registry"
	if strings.HasSuffix(subdir, "-registry") {
		id := strings.TrimSuffix(subdir, "-registry")
		if p, ok := agenttty.Get(id); ok {
			return p, true
		}
	}
	return agenttty.Provider{}, false
}
