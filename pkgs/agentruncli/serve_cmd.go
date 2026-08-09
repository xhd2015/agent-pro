package agentruncli

import (
	"context"
	"os"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/agentsend"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/tty-watch/pkgs/ttywatch"
)

func runServeSession(args []string) error {
	parsed, err := ttywatch.ParseServeArgv(args)
	if err != nil {
		return err
	}
	opts := ttywatch.ServeOptionsFromArgv(parsed)
	opts.OnListening = func(ctx context.Context, listenAddr, home, registrySubdir string) {
		startServeSendQueueDrainer(ctx, parsed.SessionID, listenAddr, home, registrySubdir)
	}
	return ttywatch.ServeSession(context.Background(), opts)
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
