package subagent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/logs"

	agentprovider "github.com/xhd2015/agent-pro/agent/cli/provider"
	"github.com/xhd2015/agent-pro/agent/cli/registry"
	"github.com/xhd2015/agent-pro/agent/event/print"
	agentexec "github.com/xhd2015/agent-pro/agent/exec"
)

func formatEventLine(line string) string {
	formatted := print.FormatTraceLine(line)
	if formatted != "" {
		return formatted
	}
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(line)), &raw); err != nil {
		return ""
	}
	eventType, _ := raw["type"].(string)
	var contentDisplay []string
	if c, ok := raw["content"].(string); ok && c != "" {
		contentDisplay = append(contentDisplay, c)
	}
	if d, ok := raw["delta"]; ok {
		if dm, ok := d.(map[string]interface{}); ok {
			if cmd, ok := dm["command"].(string); ok && cmd != "" {
				contentDisplay = append(contentDisplay, cmd)
			}
		}
	}
	if c, ok := raw["tool"].(string); ok && c != "" {
		contentDisplay = append(contentDisplay, "("+c+")")
	}
	if len(contentDisplay) > 0 {
		return eventType + ": " + strings.Join(contentDisplay, " ")
	}
	return ""
}

func runAgent(agentRunner, prompt, sessionID string, rawLog *sessionLogWriter) (string, error) {
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
		RawLog:    rawLog,
	}

	output, err := runner.Agent.Ask(context.Background(), prompt, opts, func(delta string) {
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
