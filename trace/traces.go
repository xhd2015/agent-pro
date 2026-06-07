package trace

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	agenttrace "github.com/xhd2015/agent-pro/agent_trace"
	"github.com/xhd2015/agent-pro/agent_trace/events"
)

const (
	agentTracesDirName      = "agent-traces"
	traceMetaFileName       = "metadata.json"
	tracePromptFile         = "prompt.md"
	traceLogFile            = "events.jsonl"
	traceStatusRunning      = "running"
	traceStatusStopped      = "stopped"
	traceNotRespondingTag   = "not_responding"
	traceNotRespondingAfter = 5 * time.Minute
)

type AgentTraceSession struct {
	mu       sync.Mutex
	dir      string
	metaPath string
	log      events.Logger
	meta     AgentTraceMetadata
}

func StartAgentTraceSession(dataDir string, meta AgentTraceMetadata, prompt string) (*AgentTraceSession, error) {
	root, err := agentTraceRoot(dataDir)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	now := time.Now()
	id, dir, err := reserveTraceDir(root, now)
	if err != nil {
		return nil, err
	}
	promptPath := filepath.Join(dir, tracePromptFile)
	logPath := filepath.Join(dir, traceLogFile)
	if err := os.WriteFile(promptPath, []byte(prompt), 0o644); err != nil {
		return nil, err
	}
	logFile, err := events.Open(logPath)
	if err != nil {
		return nil, err
	}
	meta.ID = id
	meta.Status = traceStatusRunning
	meta.CreatedAt = now.Format(time.RFC3339Nano)
	meta.UpdatedAt = meta.CreatedAt
	meta.PromptPath = promptPath
	meta.LogPath = logPath
	if len(meta.CommandArgs) > 0 && strings.TrimSpace(meta.CommandLine) == "" {
		meta.CommandLine = strings.Join(meta.CommandArgs, " ")
	}
	session := &AgentTraceSession{
		dir:      dir,
		metaPath: filepath.Join(dir, traceMetaFileName),
		log:      logFile,
		meta:     meta,
	}
	if err := session.saveMetaLocked(); err != nil {
		_ = logFile.Close()
		return nil, err
	}
	return session, nil
}

func ResumeAgentTraceSession(traceRef string, updates AgentTraceMetadata) (*AgentTraceSession, error) {
	dir, err := resolveTraceSessionDir(traceRef)
	if err != nil {
		return nil, err
	}
	meta, err := readAgentTraceMetadata(dir)
	if err != nil {
		return nil, err
	}
	applyAgentTraceMetadataUpdates(&meta, updates)
	now := time.Now().Format(time.RFC3339Nano)
	if strings.TrimSpace(meta.CreatedAt) == "" {
		meta.CreatedAt = now
	}
	meta.Status = traceStatusRunning
	meta.Error = ""
	meta.UpdatedAt = now
	if strings.TrimSpace(meta.PromptPath) == "" {
		meta.PromptPath = filepath.Join(dir, tracePromptFile)
	}
	if strings.TrimSpace(meta.LogPath) == "" {
		meta.LogPath = filepath.Join(dir, traceLogFile)
	}
	logFile, err := events.Open(meta.LogPath)
	if err != nil {
		return nil, err
	}
	session := &AgentTraceSession{
		dir:      dir,
		metaPath: filepath.Join(dir, traceMetaFileName),
		log:      logFile,
		meta:     meta,
	}
	if err := session.saveMetaLocked(); err != nil {
		_ = logFile.Close()
		return nil, err
	}
	return session, nil
}

func AgentTraceRoot() (string, error) {
	return agentTraceRoot("")
}

func AgentTraceRootForDataDir(dataDir string) (string, error) {
	return agentTraceRoot(dataDir)
}

func agentTraceRoot(dataDir string) (string, error) {
	root := strings.TrimSpace(dataDir)
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		root = filepath.Join(home, ".knowledge-hub")
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, agentTracesDirName), nil
}

func resolveTraceSessionDir(traceRef string) (string, error) {
	traceRef = strings.TrimSpace(traceRef)
	if traceRef == "" {
		return "", fmt.Errorf("trace session id is required")
	}
	if filepath.IsAbs(traceRef) || strings.ContainsAny(traceRef, `/\`) {
		abs, err := filepath.Abs(traceRef)
		if err != nil {
			return "", err
		}
		info, err := os.Stat(abs)
		if err != nil {
			return "", err
		}
		if !info.IsDir() {
			return "", fmt.Errorf("trace session is not a directory: %s", abs)
		}
		return filepath.Clean(abs), nil
	}
	if strings.Contains(traceRef, "..") {
		return "", fmt.Errorf("invalid trace session id")
	}
	root, err := AgentTraceRoot()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(root, traceRef)
	info, err := os.Stat(dir)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("trace session is not a directory: %s", dir)
	}
	return dir, nil
}

func applyAgentTraceMetadataUpdates(meta *AgentTraceMetadata, updates AgentTraceMetadata) {
	if meta == nil {
		return
	}
	if strings.TrimSpace(updates.Command) != "" {
		meta.Command = updates.Command
	}
	if len(updates.CommandArgs) > 0 {
		meta.CommandArgs = append([]string(nil), updates.CommandArgs...)
		if strings.TrimSpace(updates.CommandLine) == "" {
			meta.CommandLine = strings.Join(updates.CommandArgs, " ")
		}
	}
	if strings.TrimSpace(updates.CommandLine) != "" {
		meta.CommandLine = updates.CommandLine
	}
	if strings.TrimSpace(updates.TopicPath) != "" {
		meta.TopicPath = updates.TopicPath
	}
	if strings.TrimSpace(updates.Workspace) != "" {
		meta.Workspace = updates.Workspace
	}
	if strings.TrimSpace(updates.OutputPath) != "" {
		meta.OutputPath = updates.OutputPath
	}
	if strings.TrimSpace(updates.ResumeCommand) != "" {
		meta.ResumeCommand = updates.ResumeCommand
	}
	if strings.TrimSpace(updates.AgentRunnerID) != "" {
		meta.AgentRunnerID = updates.AgentRunnerID
	}
	if strings.TrimSpace(updates.Model) != "" {
		meta.Model = updates.Model
	}
}

func reserveTraceDir(root string, now time.Time) (string, string, error) {
	base := now.Format("20060102-150405.000000")
	for i := 0; i < 1000; i++ {
		id := base
		if i > 0 {
			id = fmt.Sprintf("%s-%03d", base, i+1)
		}
		dir := filepath.Join(root, id)
		err := os.Mkdir(dir, 0o755)
		if err == nil {
			return id, dir, nil
		}
		if !os.IsExist(err) {
			return "", "", err
		}
	}
	return "", "", fmt.Errorf("failed to allocate agent trace session directory under %s", root)
}

func (s *AgentTraceSession) Writer() io.Writer {
	if s == nil || s.log == nil {
		return io.Discard
	}
	return &agentTraceLogWriter{session: s}
}

func (s *AgentTraceSession) ID() string {
	if s == nil {
		return ""
	}
	return s.meta.ID
}

func (s *AgentTraceSession) Dir() string {
	if s == nil {
		return ""
	}
	return s.dir
}

func (s *AgentTraceSession) Finish(err error) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err != nil {
		s.meta.Status = "failed"
		s.meta.Error = err.Error()
	} else {
		s.meta.Status = "completed"
		s.meta.Error = ""
	}
	s.meta.UpdatedAt = time.Now().Format(time.RFC3339Nano)
	_ = s.saveMetaLocked()
	if s.log != nil {
		_ = s.log.Sync()
		_ = s.log.Close()
		s.log = nil
	}
}

func (s *AgentTraceSession) saveMetaLocked() error {
	return writeAgentTraceMetadata(s.metaPath, s.meta)
}

func writeAgentTraceMetadata(metaPath string, meta AgentTraceMetadata) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(metaPath, data, 0o644)
}

type agentTraceLogWriter struct {
	session *AgentTraceSession
}

func (w *agentTraceLogWriter) Write(p []byte) (int, error) {
	if w == nil || w.session == nil {
		return len(p), nil
	}
	w.session.mu.Lock()
	defer w.session.mu.Unlock()
	if w.session.log == nil {
		return len(p), nil
	}
	err := w.session.log.Append(p)
	if err == nil && strings.Contains(string(p), "\n") {
		_ = w.session.log.Sync()
	}
	return len(p), err
}

func loadAgentTraceSummaries(dataDir string) ([]AgentTraceSummary, error) {
	root, err := agentTraceRoot(dataDir)
	if err != nil {
		return nil, err
	}
	return loadAgentTraceSummariesFromRoot(root)
}

func loadAgentTraceSummariesFromRoot(root string) ([]AgentTraceSummary, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return []AgentTraceSummary{}, nil
		}
		return nil, err
	}
	summaries := make([]AgentTraceSummary, 0, len(entries))
	now := time.Now()
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		summary, err := loadAgentTraceSummaryFromDir(filepath.Join(root, entry.Name()), now)
		if err != nil {
			continue
		}
		summaries = append(summaries, summary)
	}
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].CreatedAt > summaries[j].CreatedAt
	})
	return summaries, nil
}

func loadAgentTraceSummaryFromDir(dir string, now time.Time) (AgentTraceSummary, error) {
	meta, err := readAgentTraceMetadataOrDefault(dir)
	if err != nil {
		return AgentTraceSummary{}, err
	}
	lineCount := countFileLines(meta.LogPath)
	meta = withAgentTraceRuntimeTags(meta, lineCount, now)
	return AgentTraceSummary{
		AgentTraceMetadata: meta,
		LogLineCount:       lineCount,
	}, nil
}

func loadAgentTraceDetail(dataDir, id string) (*AgentTraceDetail, error) {
	dir, err := agentTraceDirForIDInRoot(dataDir, id)
	if err != nil {
		return nil, err
	}
	return loadAgentTraceDetailFromDir(dir)
}

func loadAgentTraceDetailFromDir(dir string) (*AgentTraceDetail, error) {
	meta, err := readAgentTraceMetadataOrDefault(dir)
	if err != nil {
		return nil, err
	}
	promptData, err := os.ReadFile(meta.PromptPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	lines, rawLines, err := readTraceRawLines(meta.LogPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	meta = withAgentTraceRuntimeTags(meta, len(rawLines), time.Now())
	return &AgentTraceDetail{
		Metadata: meta,
		Prompt:   string(promptData),
		Messages: agenttrace.ParseMessages(lines, meta.CreatedAt),
		RawLines: rawLines,
	}, nil
}

func markAgentTraceStopped(dataDir, id string) (*AgentTraceDetail, error) {
	dir, err := agentTraceDirForIDInRoot(dataDir, id)
	if err != nil {
		return nil, err
	}
	if err := markAgentTraceDirStopped(dir); err != nil {
		return nil, err
	}
	return loadAgentTraceDetail(dataDir, id)
}

func markAgentTraceDirStopped(dir string) error {
	metaPath := filepath.Join(dir, traceMetaFileName)
	meta, err := readAgentTraceMetadata(dir)
	if err != nil {
		return err
	}
	if isTraceRunning(meta.Status) {
		meta.Status = traceStatusStopped
		meta.UpdatedAt = time.Now().Format(time.RFC3339Nano)
		meta.Tags = withoutAgentTraceRuntimeTags(meta.Tags)
		if err := writeAgentTraceMetadata(metaPath, meta); err != nil {
			return err
		}
	}
	return nil
}

func deleteAgentTraceSession(dataDir, id string) error {
	dir, err := agentTraceDirForIDInRoot(dataDir, id)
	if err != nil {
		return err
	}
	return deleteAgentTraceDir(dir)
}

func deleteAgentTraceDir(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("trace session is not a directory: %s", dir)
	}
	if _, err := readAgentTraceMetadata(dir); err != nil {
		return err
	}
	return os.RemoveAll(dir)
}

func agentTraceDirForID(id string) (string, error) {
	return agentTraceDirForIDInRoot("", id)
}

func agentTraceDirForIDInRoot(dataDir, id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" || strings.Contains(id, "..") || strings.Contains(id, "/") || strings.Contains(id, "\\") {
		return "", fmt.Errorf("invalid trace session id")
	}
	root, err := agentTraceRoot(dataDir)
	if err != nil {
		return "", err
	}
	return agentTraceDirForIDInTraceRoot(root, id)
}

func agentTraceDirForIDInTraceRoot(root, id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" || strings.Contains(id, "..") || strings.Contains(id, "/") || strings.Contains(id, "\\") {
		return "", fmt.Errorf("invalid trace session id")
	}
	return filepath.Join(root, id), nil
}

func readAgentTraceMetadata(dir string) (AgentTraceMetadata, error) {
	data, err := os.ReadFile(filepath.Join(dir, traceMetaFileName))
	if err != nil {
		return AgentTraceMetadata{}, err
	}
	var meta AgentTraceMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return AgentTraceMetadata{}, err
	}
	if meta.ID == "" {
		meta.ID = filepath.Base(dir)
	}
	return meta, nil
}

func readAgentTraceMetadataOrDefault(dir string) (AgentTraceMetadata, error) {
	meta, err := readAgentTraceMetadata(dir)
	if err == nil {
		return meta, nil
	}
	logPath := filepath.Join(dir, traceLogFile)
	info, statErr := os.Stat(logPath)
	if statErr != nil || info.IsDir() {
		return AgentTraceMetadata{}, err
	}
	createdAt := info.ModTime().Format(time.RFC3339Nano)
	return AgentTraceMetadata{
		ID:          sanitizeTraceID(filepath.Base(dir)),
		Command:     "agent-events",
		CommandLine: filepath.Base(dir),
		Status:      "completed",
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
		PromptPath:  filepath.Join(dir, tracePromptFile),
		LogPath:     logPath,
	}, nil
}

func withAgentTraceRuntimeTags(meta AgentTraceMetadata, rawLineCount int, now time.Time) AgentTraceMetadata {
	meta.Tags = withoutAgentTraceRuntimeTags(meta.Tags)
	if isAgentTraceNotResponding(meta, rawLineCount, now) {
		meta.Tags = appendAgentTraceTag(meta.Tags, traceNotRespondingTag)
	}
	return meta
}

func withoutAgentTraceRuntimeTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}
	next := make([]string, 0, len(tags))
	for _, tag := range tags {
		if tag == traceNotRespondingTag {
			continue
		}
		next = append(next, tag)
	}
	if len(next) == 0 {
		return nil
	}
	return next
}

func appendAgentTraceTag(tags []string, tag string) []string {
	for _, existing := range tags {
		if existing == tag {
			return tags
		}
	}
	return append(tags, tag)
}

func isAgentTraceNotResponding(meta AgentTraceMetadata, rawLineCount int, now time.Time) bool {
	if !isTraceRunning(meta.Status) {
		return false
	}
	lastMessageAt, ok := lastAgentTraceMessageTime(meta, rawLineCount)
	if !ok || lastMessageAt.After(now) {
		return false
	}
	return now.Sub(lastMessageAt) >= traceNotRespondingAfter
}

func lastAgentTraceMessageTime(meta AgentTraceMetadata, rawLineCount int) (time.Time, bool) {
	if rawLineCount > 0 && strings.TrimSpace(meta.LogPath) != "" {
		if info, err := os.Stat(meta.LogPath); err == nil && !info.ModTime().IsZero() {
			return info.ModTime(), true
		}
	}
	for _, value := range []string{meta.UpdatedAt, meta.CreatedAt} {
		if t, err := time.Parse(time.RFC3339Nano, value); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func readTraceRawLines(path string) ([]string, []json.RawMessage, error) {
	f, err := os.Open(path)
	if err != nil {
		return []string{}, []json.RawMessage{}, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 8*1024*1024)
	lines := []string{}
	raw := []json.RawMessage{}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		lines = append(lines, line)
		if json.Valid([]byte(line)) {
			raw = append(raw, json.RawMessage([]byte(line)))
		}
	}
	return lines, raw, scanner.Err()
}

func countFileLines(path string) int {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		count++
	}
	return count
}
