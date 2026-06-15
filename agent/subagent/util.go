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
		return "\u2014"
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

func resolveSessionID(c Config, flagSessionID string) (*sessionIDSources, error) {
	if flagSessionID != "" {
		codexID := os.Getenv("CODEX_THREAD_ID")
		return &sessionIDSources{
			sessionID:         flagSessionID,
			codexThreadID:     codexID,
			explicitSessionID: flagSessionID,
		}, nil
	}
	if v := os.Getenv(c.sessionEnvVar()); v != "" {
		codexID := os.Getenv("CODEX_THREAD_ID")
		return &sessionIDSources{
			sessionID:     v,
			codexThreadID: codexID,
			implSessionID: v,
		}, nil
	}
	if v := os.Getenv("CODEX_THREAD_ID"); v != "" {
		return &sessionIDSources{
			sessionID:     v,
			codexThreadID: v,
		}, nil
	}
	genID := generateSessionID()
	return nil, fmt.Errorf("cannot detect session id, if you're running inside opencode, try again with: `doctest agent %s --session-id %s <prompt>`, and use the same session id in subsequent followups, don't generate your session id, use the provided session id %s explicitly.", c.Cmd, genID, genID)
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
