package agentrunapi

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
)

// retargetRunnerIfNeeded activates opts.AgentRunner on the stored session when
// the requested runner name or family differs from meta.runner. switchedFamily
// is true only when the provider family changes (codex ↔ grok), not for alias
// updates (grok → grok-tty).
func retargetRunnerIfNeeded(opts Opts, meta agentstorage.SessionMeta, found bool) (agentstorage.SessionMeta, bool, error) {
	requested := strings.TrimSpace(opts.AgentRunner)
	if !found || requested == "" || opts.Store == nil {
		return meta, false, nil
	}
	stored := strings.TrimSpace(meta.Runner)
	if stored == "" {
		return meta, false, nil
	}
	if stored == requested {
		return meta, false, nil
	}
	switchedFamily := !agentstorage.SameRunnerFamily(requested, stored)
	sid := strings.TrimSpace(meta.SessionID)
	if sid == "" {
		sid = opts.SessionID
	}
	if err := opts.Store.ActivateSessionRunner(sid, requested); err != nil {
		return meta, false, err
	}
	updated, ok, err := resolveSession(opts.Store, sid)
	if err != nil {
		return meta, false, err
	}
	if !ok {
		return meta, switchedFamily, nil
	}
	if switchedFamily {
		warnRunnerSwitch(opts.Stderr, stored, requested, updated)
	}
	return updated, switchedFamily, nil
}

func warnRunnerSwitch(stderr io.Writer, from, to string, meta agentstorage.SessionMeta) {
	if stderr == nil {
		stderr = os.Stderr
	}
	fam := agentstorage.RunnerFamily(to)
	if id := strings.TrimSpace(meta.RunnerSessionID); id != "" {
		fmt.Fprintf(stderr, "warning: switching agent runner from %s to %s; resuming %s session %s\n", from, to, fam, id)
		return
	}
	fmt.Fprintf(stderr, "warning: switching agent runner from %s to %s; no existing %s session, starting fresh\n", from, to, fam)
}

func modeAfterFamilySwitch(mode Mode, meta agentstorage.SessionMeta) Mode {
	if mode != ModeSend {
		return mode
	}
	if strings.TrimSpace(meta.RunnerSessionID) != "" {
		return ModeResume
	}
	return ModeRun
}

func liveTTYFamilyMismatch(store agentstorage.Store, opts Opts, meta agentstorage.SessionMeta) bool {
	reqFam := agentstorage.RunnerFamily(effectiveRunner(opts, meta))
	if reqFam == "" {
		return false
	}
	ttyFam := liveTTYRunnerFamily(store, meta)
	if ttyFam == "" {
		return false
	}
	return ttyFam != reqFam
}

func liveTTYRunnerFamily(store agentstorage.Store, meta agentstorage.SessionMeta) string {
	if store == nil {
		return ""
	}
	termID := strings.TrimSpace(meta.TerminalSessionID)
	if termID == "" {
		termID = strings.TrimSpace(meta.SessionID)
	}
	if termID == "" {
		return ""
	}
	ttySess, err := agenttty.ResolveByTerminalID(store.Home(), termID)
	if err != nil || ttySess == nil {
		return ""
	}
	return agentstorage.RunnerFamily(ttySess.RunnerID)
}
