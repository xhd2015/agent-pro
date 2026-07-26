package subagent

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
)

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	return strconv.Quote(s)
}

func formatSessionRetryHint(c Config, sessionID, prompt string) string {
	if c.SessionRetryHint != nil {
		return c.SessionRetryHint(sessionID, prompt)
	}
	if strings.TrimSpace(prompt) != "" {
		return fmt.Sprintf("doctest agent %s --session-id %s %s", c.Cmd, sessionID, shellQuote(prompt))
	}
	return fmt.Sprintf("doctest agent %s --session-id %s <prompt>", c.Cmd, sessionID)
}

func processExists(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func parseEventTimestamp(line string) int64 {
	var event struct {
		Timestamp int64 `json:"timestamp"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(line)), &event); err != nil {
		return 0
	}
	return event.Timestamp
}

func relativeTime(timestampMs int64) string {
	if timestampMs <= 0 {
		return EmptyDisplay
	}
	now := time.Now().UnixMilli()
	diff := now - timestampMs
	if diff < 0 {
		diff = 0
	}
	seconds := diff / 1000
	if seconds < 60 {
		return fmt.Sprintf("%ds ago", seconds)
	}
	minutes := seconds / 60
	if minutes < 60 {
		return fmt.Sprintf("%dm ago", minutes)
	}
	hours := minutes / 60
	if hours < 24 {
		return fmt.Sprintf("%dh ago", hours)
	}
	days := hours / 24
	return fmt.Sprintf("%dd ago", days)
}

func generateSessionID() string {
	return "gen_" + fmt.Sprintf("%x", md5.Sum([]byte(uuid.New().String())))
}

type sessionIDSources struct {
	sessionID         string
	codexThreadID     string
	implSessionID     string
	explicitSessionID string
	agentRunner       string
}

func resolveSessionID(c Config, flagSessionID, prompt string) (*sessionIDSources, error) {
	if flagSessionID != "" {
		codexID := getenv(c, "CODEX_THREAD_ID")
		return &sessionIDSources{
			sessionID:         flagSessionID,
			codexThreadID:     codexID,
			explicitSessionID: flagSessionID,
		}, nil
	}
	if v := getenv(c, c.sessionEnvVar()); v != "" {
		codexID := getenv(c, "CODEX_THREAD_ID")
		return &sessionIDSources{
			sessionID:     v,
			codexThreadID: codexID,
			implSessionID: v,
		}, nil
	}
	if v := getenv(c, "CODEX_THREAD_ID"); v != "" {
		return &sessionIDSources{
			sessionID:     v,
			codexThreadID: v,
		}, nil
	}
	if c.AutoGenerateSessionID {
		return &sessionIDSources{
			sessionID: generateSessionID(),
		}, nil
	}
	genID := generateSessionID()
	hint := formatSessionRetryHint(c, genID, prompt)
	return nil, fmt.Errorf("cannot detect session id, if you're running inside opencode, try again with: `%s`, and use the same session id in subsequent followups, don't generate your session id, use the provided session id %s explicitly.", hint, genID)
}

func sourceLabel(srcs *sessionIDSources) string {
	if srcs.explicitSessionID != "" {
		return "--session-id"
	}
	if srcs.implSessionID != "" {
		return "AGENT_PRO_SUBAGENT_*_SESSION_ID"
	}
	if srcs.codexThreadID != "" {
		return "CODEX_THREAD_ID"
	}
	return "generated"
}

func readMockConfigSessionID(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	var cfg struct {
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return ""
	}
	return strings.TrimSpace(cfg.SessionID)
}

func resolveInnerSessionID(captureID string, isNew bool, srcs *sessionIDSources) string {
	if s := strings.TrimSpace(captureID); s != "" {
		return s
	}
	// Child runners inherit process env for FAKE_CODEX_MOCK_CONFIG; product also
	// sets it via os.Setenv before spawn. Lookup here is process-global.
	if s := readMockConfigSessionID(os.Getenv("FAKE_CODEX_MOCK_CONFIG")); s != "" {
		return s
	}
	if isNew && srcs != nil {
		return srcs.sessionID
	}
	return ""
}

func newQuestionsFile(dir string) string {
	base := time.Now().Format("2006_01_02_15_04_05")
	name := base + ".json"
	path := filepath.Join(dir, name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.WriteFile(path, nil, 0644)
		return path
	}
	for n := 1; ; n++ {
		name := base + "_" + strconv.Itoa(n) + ".json"
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			os.WriteFile(path, nil, 0644)
			return path
		}
	}
}
