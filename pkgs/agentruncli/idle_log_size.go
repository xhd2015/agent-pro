package agentruncli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
)

// idleLogSizeCache remembers the resolved Codex rollout path across Ticks.
type idleLogSizeCache struct {
	path string
}

func (c *idleLogSizeCache) fill(sample *IdleSample, home, sessionID string) {
	if sample == nil || c == nil {
		return
	}
	size, found := c.stat(home, sessionID)
	if !found {
		return
	}
	sample.LogFound = true
	sample.LogBytes = size
}

func (c *idleLogSizeCache) stat(home, sessionID string) (size int64, found bool) {
	if c.path != "" {
		st, err := os.Stat(c.path)
		if err == nil && !st.IsDir() {
			return st.Size(), true
		}
		c.path = ""
	}
	path := resolveCodexRolloutPath(home, sessionID)
	if path == "" {
		return 0, false
	}
	st, err := os.Stat(path)
	if err != nil || st.IsDir() {
		return 0, false
	}
	c.path = path
	return st.Size(), true
}

func resolveCodexRolloutPath(home, sessionID string) string {
	id, configHome := readRunnerSessionBind(home, sessionID)
	if id == "" {
		return ""
	}
	codexHome := strings.TrimSpace(configHome)
	if codexHome == "" {
		codexHome = agenttty.CodexHome()
	}
	path, ok, err := agenttty.FindCodexTranscriptBySessionID(codexHome, id)
	if err != nil || !ok {
		return ""
	}
	return path
}

func readRunnerSessionBind(home, sessionID string) (runnerSessionID, configHome string) {
	home = strings.TrimSpace(home)
	sessionID = strings.TrimSpace(sessionID)
	if home == "" || sessionID == "" {
		return "", ""
	}
	data, err := os.ReadFile(filepath.Join(home, "sessions", sessionID, "meta.json"))
	if err != nil {
		return "", ""
	}
	var meta agentstorage.SessionMeta
	if json.Unmarshal(data, &meta) != nil {
		return "", ""
	}
	return strings.TrimSpace(meta.RunnerSessionID), strings.TrimSpace(meta.AgentRunnerConfigHome)
}
