package agentruncli

import (
	"context"
	"os"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/agentsend"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/agent-pro/pkgs/tty/detection/idle"
	"github.com/xhd2015/agent-pro/pkgs/tty/detection/occupied"
	"github.com/xhd2015/tty-watch/pkgs/ttywatch"
)

func serveIdleHome(serveHome string) string {
	if home := strings.TrimSpace(os.Getenv("AGENT_RUN_HOME")); home != "" {
		return home
	}
	return strings.TrimSpace(serveHome)
}

// startServeIdleWatchdog arms the keep-alive idle-exit loop when idle-policy.json
// says exit_on_idle. SoftExit injects /exit; Shutdown cancels the serve ctx.
// Detection is runner-agnostic: resting snapshot unchanged + occupy space probe.
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
	runner := strings.TrimSpace(provider.ID)

	inject := func(text string) error {
		if runner == "" {
			return errIdleNoRunner
		}
		return agenttty.InjectMessage(listenAddr, sessionID, runner, text, false)
	}
	syncSnap := func() (string, error) {
		return ttywatch.SnapshotText(listenAddr, sessionID)
	}

	w := idle.New(found, idle.Policy{ExitOnIdle: p.ExitOnIdle, IdleTimeout: p.IdleTimeout}, idle.Watchdog{
		Snapshot: syncSnap,
		ProbeOccupied: func() occupied.Status {
			return occupied.Probe(occupied.IO{
				Snapshot: syncSnap,
				Inject:   inject,
			})
		},
		Inject: inject,
		Ready: func(snapshot string) bool {
			if provider.CheckWritable == nil {
				return true
			}
			return provider.CheckWritable([]byte(snapshot)).Ready
		},
		QueueLen: func() int {
			if runner == "" {
				return 0
			}
			n, err := agentsend.QueueLen(home, runner, sessionID)
			if err != nil {
				return 0
			}
			return n
		},
		SoftExit: func() {
			if runner == "" {
				return
			}
			// Injection has its own readiness retries. Do not let those retries
			// postpone the watchdog grace deadline: the hard shutdown is the
			// bounded fallback when a TUI does not consume /exit.
			go func() {
				_ = agenttty.InjectMessage(listenAddr, sessionID, runner, "/exit", true)
			}()
		},
		Shutdown: func() {
			if cancel != nil {
				cancel()
			}
		},
	})
	go idle.RunLoop(ctx, w)
}

type idleRunnerError string

func (e idleRunnerError) Error() string { return string(e) }

const errIdleNoRunner = idleRunnerError("idle: empty runner id")
