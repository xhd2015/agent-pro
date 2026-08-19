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

// sleepCtx sleeps d or returns false if ctx is done. d<=0 is a no-op.
func sleepCtx(ctx context.Context, d time.Duration) bool {
	if ctx.Err() != nil {
		return false
	}
	if d <= 0 {
		return true
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}

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
	for ctx.Err() == nil {
		first, gap := IdleWatchSchedule(w.Timeout)
		if !sleepCtx(ctx, first) {
			return
		}
		w.Tick()
		if w.idleHits == 0 {
			continue
		}
		if !sleepCtx(ctx, gap) {
			return
		}
		w.Tick()
		if w.idleHits == 0 {
			continue
		}
		if !sleepCtx(ctx, gap) {
			return
		}
		w.Tick()
		if !w.softDone {
			continue
		}
		if !sleepCtx(ctx, w.Grace) {
			return
		}
		w.forceShutdown()
		return
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
	if provider.ID == "codex-tty" && sample.Sendable && sample.Screen == "idle" {
		sample.InputBox = probeCodexOccupancy(listenAddr, sessionID, provider.ID, text)
	} else {
		sample.InputBox = agenttty.DetectInputBox(text).String()
	}
	if n, qerr := agentsend.QueueLen(home, provider.ID, sessionID); qerr == nil {
		sample.QueueLen = n
	}
	return sample
}

const (
	codexOccupancyProbeCap  = 5 * time.Second
	codexOccupancyPollEvery = 200 * time.Millisecond
)

// probeCodexOccupancy types a space and watches the last-› remainder.
// Placeholder collapses to empty/whitespace; a real draft stays nonempty.
// Always undoes with one DEL. Only called when sendable+screen idle.
func probeCodexOccupancy(listenAddr, sessionID, runner, text string) string {
	before := agenttty.LastComposerRemainder(text)
	if strings.TrimSpace(before) == "" {
		return agenttty.InputBoxEmpty.String()
	}
	if err := agenttty.InjectMessage(listenAddr, sessionID, runner, " ", false); err != nil {
		return agenttty.DetectInputBox(text).String()
	}
	defer func() {
		_ = agenttty.InjectMessage(listenAddr, sessionID, runner, "\x7f", false)
	}()
	deadline := time.Now().Add(codexOccupancyProbeCap)
	for {
		snap, err := ttywatch.SnapshotText(listenAddr, sessionID)
		if err == nil {
			after := agenttty.LastComposerRemainder(snap)
			if after != before {
				if strings.TrimSpace(after) == "" {
					return agenttty.InputBoxEmpty.String()
				}
				return agenttty.InputBoxOccupied.String()
			}
		}
		if !time.Now().Before(deadline) {
			return agenttty.InputBoxOccupied.String()
		}
		time.Sleep(codexOccupancyPollEvery)
	}
}
