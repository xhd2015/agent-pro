package sessionstate

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/xhd2015/agent-pro/agent/agentui/runner"
	"github.com/xhd2015/agent-pro/agent/session"
)

type Meta struct {
	SessionID         string `json:"session_id"`
	OpencodeSessionID string `json:"opencode_session_id,omitempty"`
	Feature           string `json:"feature"`
	Model             string `json:"model"`
}

func ReadFromDir(dir, sessionID string) (string, string, string, string, []string) {
	var meta Meta
	if err := session.ReadJSON(dir, "metadata.json", &meta); err != nil {
		return "", "", "", "", nil
	}
	if meta.SessionID == "" {
		meta.SessionID = sessionID
	}

	lines, err := session.ReadLines(dir, "events.jsonl")
	if err != nil {
		return meta.SessionID, meta.OpencodeSessionID, meta.Feature, meta.Model, nil
	}

	var logs []string
	for _, line := range lines {
		formatted := runner.FormatLogLine(line)
		if formatted != "" {
			logs = append(logs, formatted)
		}
	}

	return meta.SessionID, meta.OpencodeSessionID, meta.Feature, meta.Model, logs
}

func Resolve(agentName, resumeID string) (sessionID, opencodeSessionID, sessionDir, feature, llmModel string, logs []string, err error) {
	if resumeID == "" {
		return "", "", "", "", "", nil, nil
	}
	dir, err := session.Dir(agentName, resumeID)
	if err != nil {
		return "", "", "", "", "", nil, fmt.Errorf("resume session: %w", err)
	}
	sid, osid, feat, model, eventLogs := ReadFromDir(dir, resumeID)
	if sid == "" {
		return "", "", "", "", "", nil, fmt.Errorf("session not found: %s (check the session ID)", resumeID)
	}
	return sid, osid, dir, feat, model, eventLogs, nil
}

func NewID(prefix string) string {
	var b [8]byte
	rand.Read(b[:])
	return prefix + hex.EncodeToString(b[:])
}

func WriteMeta(dir string, meta Meta) error {
	return session.WriteJSON(dir, "metadata.json", meta)
}

func UpdateOpencodeSessionID(dir, opencodeSessionID string) {
	var meta Meta
	if session.ReadJSON(dir, "metadata.json", &meta) == nil {
		meta.OpencodeSessionID = opencodeSessionID
		session.WriteJSON(dir, "metadata.json", meta)
	}
}
