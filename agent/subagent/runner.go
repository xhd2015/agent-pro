package subagent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/logs"

	"github.com/xhd2015/agent-pro/agent/event/print"
	eventtypes "github.com/xhd2015/agent-pro/agent/event/types"
	agentprovider "github.com/xhd2015/agent-pro/agent/cli/provider"
	"github.com/xhd2015/agent-pro/agent/cli/registry"
	agentexec "github.com/xhd2015/agent-pro/agent/exec"
)

func formatEventLine(line string) string {
	trimmed := strings.TrimSpace(line)
	if trimmed != "" && strings.HasPrefix(trimmed, "{") {
		var event eventtypes.AgentEvent
		if err := json.Unmarshal([]byte(trimmed), &event); err == nil && isFormatEventAgentType(event.Type) {
			return print.FormatAgentEvent(event)
		}
	}
	return print.FormatTraceLine(line)
}

func isFormatEventAgentType(t eventtypes.ActionType) bool {
	switch t {
	case eventtypes.ActionThink, eventtypes.ActionToolCall, eventtypes.ActionMessage,
		eventtypes.ActionError, eventtypes.ActionDone,
		eventtypes.ActionStepStart, eventtypes.ActionStepFinish,
		eventtypes.ActionSleep:
		return true
	}
	return false
}

func runAgent(ctx context.Context, agentRunner, model, prompt, sessionID string, rawLog *sessionLogWriter, stdout io.Writer) (string, error) {
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	env := agentexec.NewEnv(&agentexec.PathsConfig{
		RootDirName: ".agent-pro",
		DataDirName: "data",
		BinDirName:  "bin",
	}, "AGENT_PRO_CONFIG_HOME")

	runner, err := agentprovider.Build(registry.AgentRunnerID(agentRunner), "", ".", env)
	if err != nil {
		return "", err
	}

	opts := &registry.AskOptions{
		Workspace: ".",
		SessionID: sessionID,
		Model:     model,
		RawLog:    rawLog,
	}

	output, err := runner.Agent.Ask(ctx, prompt, opts, func(delta string) {
		writeStdout(stdout, delta)
	})
	if err != nil {
		return output, err
	}
	return output, nil
}

func traceSession(c Config, opts Options) error {
	srcs, err := resolveSessionID(c, opts.SessionID, opts.Prompt)
	if err != nil {
		return err
	}

	var sessionDir string
	if opts.SessionLayout.flatDir() {
		sessionDir = opts.SessionLayout.Dir
	} else {
		base, baseErr := sessionsBase(c, opts)
		if baseErr != nil {
			return baseErr
		}
		var found bool
		sessionDir, found = findSession(c, base, srcs.sessionID, srcs)
		if !found || sessionDir == "" {
			sessionDir, _, err = findOrCreateSession(c, opts, srcs.sessionID, srcs)
			if err != nil {
				return fmt.Errorf("session: %w", err)
			}
		}
	}

	paths := resolvedSessionPaths(sessionDir, opts.SessionLayout)
	eventsPath := paths.eventsPath

	printHeader := func(eventCount int) {
		FprintTraceHeader(os.Stdout, srcs.sessionID, eventCount)
	}

	data, err := os.ReadFile(eventsPath)
	hasEvents := err == nil
	if err != nil {
		if os.IsNotExist(err) {
			printHeader(0)
			Logf("(no events yet)")
		} else {
			return fmt.Errorf("read events.jsonl: %w", err)
		}
	} else {
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		eventCount := 0
		for _, line := range lines {
			if line != "" {
				eventCount++
			}
		}
		printHeader(eventCount)

		n := 0
		var state print.FormatState
		for _, line := range lines {
			if line == "" {
				continue
			}
			header, body, isMsg := state.FormatLine(line)
			if header == "" && body == "" && !isMsg {
				continue
			}
			if isMsg {
				if header != "" {
					n++
					Logf("[%d]  %s", n, header)
				}
				fmt.Print(body)
			} else {
				n++
				Logf("[%d]  %s", n, header)
				if body != "" {
					fmt.Print(body)
				}
			}
		}
		state.Flush()
	}

	sessionLive := isSessionLiveAt(paths.pidPath)
	if !sessionLive || !hasEvents {
		FprintTraceFooterFrame(os.Stdout)
		Logf("Done (session finished)")
		FprintlnTraceFooterRule(os.Stdout)
		return nil
	}

	Logf("Following new events (Ctrl+C to stop)...")

	var n int
	if data != nil {
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		var preState print.FormatState
		for _, line := range lines {
			if line != "" {
				header, _, isMsg := preState.FormatLine(line)
				if (isMsg && header != "") || (!isMsg && header != "") {
					n++
				}
			}
		}
		preState.Flush()
	}
	n = n + 1

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var watchErr error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		var watchState print.FormatState
		watchErr = logs.WatchLine(ctx, eventsPath, logs.WatchLineOptions{}, func(line string) error {
			header, body, isMsg := watchState.FormatLine(line)
			if header == "" && body == "" && !isMsg {
				return nil
			}
			if isMsg {
				if header != "" {
					n++
					Logf("[%d]  %s", n, header)
				}
				fmt.Print(body)
			} else {
				n++
				Logf("[%d]  %s", n, header)
				if body != "" {
					fmt.Print(body)
				}
			}
			return nil
		})
	}()

	for {
		time.Sleep(2 * time.Second)
		if !isSessionLiveAt(paths.pidPath) {
			cancel()
			break
		}
	}

	wg.Wait()

	if watchErr != nil && watchErr != context.Canceled {
		return watchErr
	}

	FprintTraceFooterFrame(os.Stdout)
	Logf("Session finished")
	FprintlnTraceFooterRule(os.Stdout)

	return nil
}
