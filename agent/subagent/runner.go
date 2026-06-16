package subagent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/logs"

	"github.com/xhd2015/agent-pro/agent/event/print"
	agentprovider "github.com/xhd2015/agent-pro/agent/cli/provider"
	"github.com/xhd2015/agent-pro/agent/cli/registry"
	agentexec "github.com/xhd2015/agent-pro/agent/exec"
)

func formatEventLine(line string) string {
	return print.FormatTraceLine(line)
}

func runAgent(ctx context.Context, agentRunner, model, prompt, sessionID string, rawLog *sessionLogWriter) (string, error) {
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
		fmt.Print(delta)
	})
	if err != nil {
		return output, err
	}
	return output, nil
}

func traceSession(c Config, opts Options) error {
	srcs, err := resolveSessionID(c, opts.SessionID)
	if err != nil {
		return err
	}

	sessionDir, _, err := findOrCreateSession(c, opts, srcs.sessionID, srcs)
	if err != nil {
		return fmt.Errorf("session: %w", err)
	}

	eventsPath := filepath.Join(sessionDir, "events.jsonl")

	printHeader := func(eventCount int) {
		fmt.Fprintf(os.Stdout, "\n\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\n")
		fmt.Fprintf(os.Stdout, "  Session: %s\n", srcs.sessionID)
		fmt.Fprintf(os.Stdout, "  Events:  %d lines\n", eventCount)
		fmt.Fprintf(os.Stdout, "\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\n\n")
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
		for _, line := range lines {
			if line == "" {
				continue
			}
			formatted := formatEventLine(line)
			if formatted != "" {
				n++
				Logf("[%d]  %s", n, formatted)
			}
		}
	}

	sessionLive := isSessionLive(sessionDir)
	if !sessionLive || !hasEvents {
		fmt.Fprintf(os.Stdout, "\n\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\n")
		Logf("Done (session finished)")
		fmt.Fprintf(os.Stdout, "\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\n")
		return nil
	}

	Logf("Following new events (Ctrl+C to stop)...")

	var n int
	if data != nil {
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		for _, line := range lines {
			if line != "" && formatEventLine(line) != "" {
				n++
			}
		}
	}
	n = n + 1

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var watchErr error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		watchErr = logs.WatchLine(ctx, eventsPath, logs.WatchLineOptions{}, func(line string) error {
			formatted := formatEventLine(line)
			if formatted != "" {
				n++
				Logf("[%d]  %s", n, formatted)
			}
			return nil
		})
	}()

	for {
		time.Sleep(2 * time.Second)
		if !isSessionLive(sessionDir) {
			cancel()
			break
		}
	}

	wg.Wait()

	if watchErr != nil && watchErr != context.Canceled {
		return watchErr
	}

	fmt.Fprintf(os.Stdout, "\n\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\n")
	Logf("Session finished")
	fmt.Fprintf(os.Stdout, "\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\n")

	return nil
}
