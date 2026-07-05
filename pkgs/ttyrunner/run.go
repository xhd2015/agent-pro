package ttyrunner

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/pkgs/groktty"
)

// RunOptions configures a unified TTY runner invocation.
type RunOptions struct {
	Home              string
	Workspace         string
	Prompt            string
	Model             string
	ResumeSessionID   string
	RunnerID          string
	AgentSessionID    string
	SettingsPath      string
	AgentPath               string
	AgentRunnerConfigHome   string
	KeepTerminalAlive       bool
	Stderr            io.Writer
	Emit              func(types.AgentEvent) error
	OnTerminalSessionID func(string)
}

// Run starts a TTY session via groktty engine with provider hooks and tty.json dual-write.
func Run(ctx context.Context, opts RunOptions) (captured, runnerSessionID string, err error) {
	ensureStubRegistered()
	runnerID := strings.TrimSpace(opts.RunnerID)
	if runnerID == "" {
		runnerID = "grok-tty"
	}
	provider, ok := Get(runnerID)
	if !ok {
		return "", "", fmt.Errorf("unknown TTY runner: %s", runnerID)
	}

	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}

	terminalSessionID := ""
	onTerminalID := func(id string) {
		terminalSessionID = strings.TrimSpace(id)
		if opts.OnTerminalSessionID != nil {
			opts.OnTerminalSessionID(terminalSessionID)
		}
	}

	keepBlocking := false
	if runnerID == "stub-tty" && opts.KeepTerminalAlive {
		if sc := loadStubScenario(); sc != nil && !sc.ExitAfterTurn {
			keepBlocking = true
		}
	}
	grokOpts := groktty.RunOptions{
		Home:                  opts.Home,
		Workspace:             opts.Workspace,
		Prompt:                opts.Prompt,
		Model:                 opts.Model,
		ResumeSessionID:       opts.ResumeSessionID,
		SettingsPath:          opts.SettingsPath,
		AgentPath:               opts.AgentPath,
		AgentRunnerConfigHome:   opts.AgentRunnerConfigHome,
		RunnerID:                runnerID,
		StderrPrefix:          runnerID,
		RegistryDir:           provider.RegistryDir,
		BannerProvider:        provider.BannerProvider,
		BannerMarkers:         provider.BannerMarkers,
		DisableTail:           provider.DisableTail,
		KeepTerminalAlive:     opts.KeepTerminalAlive,
		KeepTerminalBlocking:  keepBlocking,
		BuildArgv:             provider.BuildArgv,
		Stderr:                opts.Stderr,
		Emit:                  opts.Emit,
		OnTerminalSessionID:   onTerminalID,
	}

	var tailWG sync.WaitGroup
	var tailCancel context.CancelFunc
	var tailRunnerSessionID string
	var tailMu sync.Mutex

	if provider.StartEventTail != nil && opts.Emit != nil && runnerID == "stub-tty" {
		_, cancel := context.WithCancel(ctx)
		tailCancel = cancel
		tailWG.Add(1)
		go func() {
			defer tailWG.Done()
			id, tailErr := provider.StartEventTail(TailContext{
				ScenarioPath: os.Getenv("AGENT_RUN_STUB_TTY_SCENARIO"),
				Emit:         opts.Emit,
			})
			if tailErr == nil {
				tailMu.Lock()
				tailRunnerSessionID = id
				tailMu.Unlock()
			}
		}()
	}

	captured, grokSessionID, err := groktty.Run(ctx, grokOpts)
	if tailCancel != nil {
		tailCancel()
		tailWG.Wait()
	}

	if terminalSessionID != "" && opts.AgentSessionID != "" {
		entry, readErr := groktty.ReadRegistryFor(opts.Home, provider.RegistryDir, terminalSessionID, runnerID)
		if readErr == nil && entry != nil {
			screenStatus := "unknown"
			if provider.DetectScreenStatus != nil {
				screenStatus = provider.DetectScreenStatus(nil)
			}
			_ = WriteTTYJSON(opts.Home, TTYSnapshot{
				RunnerID:          runnerID,
				AgentSessionID:    opts.AgentSessionID,
				TerminalSessionID: terminalSessionID,
				ListenAddr:        entry.ListenAddr,
				PID:               entry.PID,
				CreatedAt:         entry.CreatedAt,
				ScreenStatus:      screenStatus,
				Alive:             opts.KeepTerminalAlive || err == nil,
			})
		}
	}

	runnerSessionID = grokSessionID
	if strings.TrimSpace(tailRunnerSessionID) != "" {
		runnerSessionID = tailRunnerSessionID
	}
	return captured, runnerSessionID, err
}

// WriteTTYJSONOnStart writes tty.json when registry entry is known (called from run hook).
func WriteTTYJSONOnStart(home, runnerID, agentSessionID string, entry RegistryEntry) error {
	provider, _ := Get(runnerID)
	screenStatus := "unknown"
	if provider.DetectScreenStatus != nil {
		screenStatus = provider.DetectScreenStatus(nil)
	}
	return WriteTTYJSON(home, TTYSnapshot{
		RunnerID:          runnerID,
		AgentSessionID:    agentSessionID,
		TerminalSessionID: entry.SessionID,
		ListenAddr:        entry.ListenAddr,
		PID:               entry.PID,
		CreatedAt:         entry.CreatedAt,
		ScreenStatus:      screenStatus,
		Alive:             true,
	})
}

// DualWriteAfterRegistry should be called when registry is written during run.
// groktty.Run writes registry internally; we patch via OnTerminalSessionID in agentui.

func dualWriteAfterRegistry(home, runnerID, agentSessionID, terminalSessionID string) {
	if agentSessionID == "" || terminalSessionID == "" {
		return
	}
	provider, _ := Get(runnerID)
	dir := runnerID + "-registry"
	if provider.RegistryDir != "" {
		dir = provider.RegistryDir
	}
	for i := 0; i < 20; i++ {
		entry, err := groktty.ReadRegistryFor(home, dir, terminalSessionID, runnerID)
		if err == nil && entry != nil {
			_ = WriteTTYJSONOnStart(home, runnerID, agentSessionID, *entry)
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
}

// PatchRunWithDualWrite wraps agentui callback to dual-write tty.json at terminal id assignment.
func PatchRunWithDualWrite(home, runnerID, agentSessionID string, onID func(string)) func(string) {
	return func(id string) {
		id = strings.TrimSpace(id)
		if onID != nil {
			onID(id)
		}
		if id != "" {
			go dualWriteAfterRegistry(home, runnerID, agentSessionID, id)
		}
	}
}