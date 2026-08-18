package agentruncli

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentsend"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/tty-watch/pkgs/ttywatch"
)

const idleWatchPollInterval = 2 * time.Second

func serveIdleHome(serveHome string) string {
	if home := strings.TrimSpace(os.Getenv("AGENT_RUN_HOME")); home != "" {
		return home
	}
	return strings.TrimSpace(serveHome)
}

// startServeIdleWatchdog arms the keep-alive idle-exit loop when idle-policy.json
// says exit_on_idle. SoftExit injects /exit; Shutdown cancels the serve ctx.
func startServeIdleWatchdog(ctx context.Context, cancel context.CancelFunc, sessionID, listenAddr, serveHome, registrySubdir string) {
	home := serveIdleHome(serveHome)
	sessionID = strings.TrimSpace(sessionID)
	if home == "" || sessionID == "" || listenAddr == "" {
		return
	}
	p, found, err := ReadIdlePolicy(home, sessionID)
	if err != nil || !found || !p.ExitOnIdle {
		return
	}
	provider, _ := providerForRegistrySubdir(registrySubdir)
	w := NewIdleWatchdog(found, p, IdleWatchdog{
		Sample: func() IdleSample {
			return sampleServeIdle(listenAddr, sessionID, home, provider)
		},
		SoftExit: func() {
			runner := strings.TrimSpace(provider.ID)
			if runner == "" {
				return
			}
			_ = agenttty.InjectMessage(listenAddr, sessionID, runner, "/exit", true)
		},
		Shutdown: func() {
			if cancel != nil {
				cancel()
			}
		},
	})
	go runIdleWatchLoop(ctx, w)
}

func runIdleWatchLoop(ctx context.Context, w *IdleWatchdog) {
	if w == nil {
		return
	}
	ticker := time.NewTicker(idleWatchPollInterval)
	defer ticker.Stop()
	w.Tick()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.Tick()
		}
	}
}

func sampleServeIdle(listenAddr, sessionID, home string, provider agenttty.Provider) IdleSample {
	sample := IdleSample{
		Screen:   "unknown",
		InputBox: "unknown",
	}
	text, err := ttywatch.SnapshotText(listenAddr, sessionID)
	if err != nil {
		return sample
	}
	scrollback := []byte(text)
	if provider.CheckWritable != nil {
		sample.Sendable = provider.CheckWritable(scrollback).Ready
	}
	if provider.DetectScreenStatus != nil {
		if screen := strings.TrimSpace(provider.DetectScreenStatus(scrollback)); screen != "" {
			sample.Screen = screen
		}
	}
	sample.InputBox = agenttty.DetectInputBox(text).String()
	if n, qerr := agentsend.QueueLen(home, provider.ID, sessionID); qerr == nil {
		sample.QueueLen = n
	}
	return sample
}
