package subagent

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// SessionLayout overrides where subagent reads/writes session files.
// Zero value = default layout under ~/.agent-pro/subagent/<role>/sessions/.
type SessionLayout struct {
	// Dir is the session root. When set, use this directory directly
	// (no date nesting, no sess_* auto-create). Required for task-hub integration.
	Dir string

	// Per-file overrides. Empty = filepath.Join(Dir, <default-name>).
	MetaPath     string
	MessagesPath string // empty = skip messages.jsonl writes
	EventsPath   string
	PIDPath      string
	QuestionsDir string // custom path when questions enabled
	ProgressDir  string // custom path when progress enabled

	// Optional features — default ON in legacy mode (Dir unset).
	// When Dir is set, caller must set these explicitly.
	QuestionsEnabled bool
	ProgressEnabled  bool
}

func (l SessionLayout) flatDir() bool {
	return l.Dir != ""
}

func (l SessionLayout) questionsEnabled() bool {
	if !l.flatDir() {
		return true
	}
	return l.QuestionsEnabled
}

func (l SessionLayout) progressEnabled() bool {
	if !l.flatDir() {
		return true
	}
	return l.ProgressEnabled
}

func (l SessionLayout) metaPath() string {
	if l.MetaPath != "" {
		return l.MetaPath
	}
	if l.Dir != "" {
		return filepath.Join(l.Dir, "meta.json")
	}
	return "meta.json"
}

func (l SessionLayout) messagesPath() string {
	if l.MessagesPath != "" {
		if l.Dir != "" && !filepath.IsAbs(l.MessagesPath) {
			return filepath.Join(l.Dir, l.MessagesPath)
		}
		return l.MessagesPath
	}
	if l.Dir != "" {
		return filepath.Join(l.Dir, "messages.jsonl")
	}
	return "messages.jsonl"
}

func (l SessionLayout) skipMessages() bool {
	return l.flatDir() && l.MessagesPath == ""
}

func (l SessionLayout) eventsPath() string {
	if l.EventsPath != "" {
		return l.EventsPath
	}
	if l.Dir != "" {
		return filepath.Join(l.Dir, "events.jsonl")
	}
	return "events.jsonl"
}

func (l SessionLayout) pidPath() string {
	if l.PIDPath != "" {
		return l.PIDPath
	}
	if l.Dir != "" {
		return filepath.Join(l.Dir, "pid")
	}
	return "pid"
}

func (l SessionLayout) questionsDir() string {
	if l.QuestionsDir != "" {
		return l.QuestionsDir
	}
	if l.Dir != "" {
		return filepath.Join(l.Dir, "questions")
	}
	return "questions"
}

func (l SessionLayout) progressDir() string {
	if l.ProgressDir != "" {
		return l.ProgressDir
	}
	if l.Dir != "" {
		return filepath.Join(l.Dir, "progress")
	}
	return "progress"
}

func metaExplicitSessionID(m map[string]any) string {
	if s, _ := m["explicit_session_id"].(string); s != "" {
		return s
	}
	if s, _ := m["agent_session_id"].(string); s != "" {
		return s
	}
	return ""
}

func metaMatchesSessionID(m map[string]any, threadID, matchField string) bool {
	if v, ok := m[matchField]; ok {
		if s, ok := v.(string); ok && s == threadID {
			return true
		}
	}
	if metaExplicitSessionID(m) == threadID {
		return true
	}
	return false
}

type resolvedPaths struct {
	metaPath     string
	messagesPath string
	eventsPath   string
	pidPath      string
	questionsDir string
	progressDir  string
}

func resolvedSessionPaths(sessionDir string, layout SessionLayout) resolvedPaths {
	if layout.flatDir() {
		return resolvedPaths{
			metaPath:     layout.metaPath(),
			messagesPath: layout.messagesPath(),
			eventsPath:   layout.eventsPath(),
			pidPath:      layout.pidPath(),
			questionsDir: layout.questionsDir(),
			progressDir:  layout.progressDir(),
		}
	}
	return resolvedPaths{
		metaPath:     filepath.Join(sessionDir, "meta.json"),
		messagesPath: filepath.Join(sessionDir, "messages.jsonl"),
		eventsPath:   filepath.Join(sessionDir, "events.jsonl"),
		pidPath:      filepath.Join(sessionDir, "pid"),
		questionsDir: filepath.Join(sessionDir, "questions"),
		progressDir:  filepath.Join(sessionDir, "progress"),
	}
}

func readMetaMap(metaPath string) (map[string]any, error) {
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}