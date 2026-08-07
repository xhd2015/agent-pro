package agentrunapi

import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

func TestComputeRunnerExited_codexFooterBeatsSendableYes(t *testing.T) {
	// Residual glyphs would set sendable=yes without the writable fix; exit
	// markers must still classify as exited (belt-and-suspenders).
	layers := lifecycleLayers{
		processStatus:      "alive",
		processPID:         12345,
		terminalStatus:     "reachable",
		terminalSendable:   "yes",
		runnerStatus:       "bound",
		scrollback:         "› leftover\nTo continue this session, run codex resume 019fdca1-3893-7fa3-a8aa-ebc1ccc750a0\n",
	}
	meta := agentstorage.SessionMeta{Runner: "codex-tty", Status: "running"}
	got := computeRunnerExited(layers, meta)
	if got == nil || !*got {
		t.Fatalf("want exited=true with Codex footer even if sendable=yes, got %v", ptrBoolString(got))
	}
}

func TestComputeRunnerExited_codexStandaloneResumeNotExit(t *testing.T) {
	layers := lifecycleLayers{
		processStatus:    "alive",
		processPID:       1,
		terminalStatus:   "reachable",
		terminalSendable: "yes",
		runnerStatus:     "bound",
		scrollback:       "» mention codex resume 019fdca1-3893-7fa3-a8aa-ebc1ccc750a0 in docs\n",
	}
	meta := agentstorage.SessionMeta{Runner: "codex-tty"}
	got := computeRunnerExited(layers, meta)
	if got == nil || *got {
		t.Fatalf("resume cmd alone + sendable=yes should be live, got %v", ptrBoolString(got))
	}
}

func TestComputeRunnerExited_codexFooterNotAppliedToGrok(t *testing.T) {
	layers := lifecycleLayers{
		processStatus:    "alive",
		processPID:       1,
		terminalStatus:   "reachable",
		terminalSendable: "yes",
		runnerStatus:     "bound",
		scrollback:       "To continue this session, run codex resume 019fdca1-3893-7fa3-a8aa-ebc1ccc750a0\n",
	}
	meta := agentstorage.SessionMeta{Runner: "grok-tty"}
	got := computeRunnerExited(layers, meta)
	if got == nil || *got {
		t.Fatalf("Codex footer on grok-tty must not force exited, got %v", ptrBoolString(got))
	}
}

func TestComputeRunnerExited_terminalExitedMarker(t *testing.T) {
	layers := lifecycleLayers{
		processStatus:    "alive",
		processPID:       1,
		terminalStatus:   "reachable",
		terminalSendable: "no",
		runnerStatus:     "bound",
		scrollback:       "done\n[Terminal exited]\n",
	}
	meta := agentstorage.SessionMeta{Runner: "codex-tty"}
	got := computeRunnerExited(layers, meta)
	if got == nil || !*got {
		t.Fatalf("want exited=true for [Terminal exited], got %v", ptrBoolString(got))
	}
}

func TestComputeRunnerExited_commandExitedPrimary(t *testing.T) {
	// Keep-alive serve still "alive" + sendable residual, but command_exited.
	layers := lifecycleLayers{
		processStatus:    "alive",
		processPID:       99,
		commandPID:       100,
		commandExited:    true,
		commandAlive:     boolPtr(false),
		terminalStatus:   "reachable",
		terminalSendable: "yes",
		runnerStatus:     "bound",
		scrollback:       "› leftover",
	}
	meta := agentstorage.SessionMeta{Runner: "codex-tty"}
	got := computeRunnerExited(layers, meta)
	if got == nil || !*got {
		t.Fatalf("command_exited should force exited=true, got %v", ptrBoolString(got))
	}
}

func TestComputeRunnerExited_commandPIDDead(t *testing.T) {
	layers := lifecycleLayers{
		processStatus:    "alive",
		processPID:       99,
		commandPID:       100,
		commandAlive:     boolPtr(false),
		terminalStatus:   "reachable",
		terminalSendable: "yes",
		runnerStatus:     "bound",
		scrollback:       "› leftover",
	}
	meta := agentstorage.SessionMeta{Runner: "codex-tty"}
	got := computeRunnerExited(layers, meta)
	if got == nil || !*got {
		t.Fatalf("dead command_pid should force exited=true, got %v", ptrBoolString(got))
	}
}

func TestComputeRunnerExited_commandPIDAliveIsLive(t *testing.T) {
	layers := lifecycleLayers{
		processStatus:    "alive",
		processPID:       99,
		commandPID:       100,
		commandAlive:     boolPtr(true),
		terminalStatus:   "reachable",
		terminalSendable: "no",
		runnerStatus:     "bound",
		// Footer alone must not override a live command_pid.
		scrollback: "To continue this session, run codex resume 019fdca1-3893-7fa3-a8aa-ebc1ccc750a0\n",
	}
	meta := agentstorage.SessionMeta{Runner: "codex-tty"}
	got := computeRunnerExited(layers, meta)
	if got == nil || *got {
		t.Fatalf("live command_pid should be live even with footer, got %v", ptrBoolString(got))
	}
}

func boolPtr(v bool) *bool { return &v }

func ptrBoolString(p *bool) string {
	if p == nil {
		return "nil"
	}
	if *p {
		return "true"
	}
	return "false"
}
