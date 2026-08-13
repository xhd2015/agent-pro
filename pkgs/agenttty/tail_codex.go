package agenttty

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

// codexResumeRe matches "codex resume <id>" and "codex --resume <id>".
var codexResumeRe = regexp.MustCompile(`\bcodex\s+(?:--)?resume\s+([0-9a-fA-F-]{20,})\b`)

// CodexHome returns the codex data directory.
func CodexHome() string {
	if home := strings.TrimSpace(os.Getenv("CODEX_HOME")); home != "" {
		return home
	}
	if home := strings.TrimSpace(os.Getenv("HOME")); home != "" {
		return filepath.Join(home, ".codex")
	}
	return ".codex"
}

// FindCodexResumeSessionID extracts a codex resume session id from terminal text.
func FindCodexResumeSessionID(text string) string {
	matches := codexResumeRe.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return ""
	}
	return matches[len(matches)-1][1]
}

// openCodexBindBudget is how long --open/--detach waits for a Codex session id
// after inject so meta.runner_session_id can be bound before the caller returns.
// Attach-instant and OpenCloseExits mean we cannot rely on post-exit footers.
const openCodexBindBudget = 20 * time.Second

// openCodexBindPoll is the poll interval while waiting for a rollout / footer.
const openCodexBindPoll = 200 * time.Millisecond

// DiscoverCodexSessionID tries one-shot discovery of a codex session for workspace
// after runStart. prompt is the agent-run open inject text (may be empty).
//
// Bind policy (concurrent same-cwd safe):
//  1. Prefer this TTY's resume footer when present (per-terminal).
//  2. Else disk scan: cwd + time; when prompt is set, first real user message
//     must match (skip Codex <environment_context> blobs).
//  3. Never bind "newest cwd" when multiple candidates remain (fail closed).
//
// Empty scrollback skips footer extraction.
func DiscoverCodexSessionID(codexHome, workspace string, runStart time.Time, scrollback, prompt string) (sessionID string, ok bool) {
	codexHome = strings.TrimSpace(codexHome)
	if codexHome == "" {
		codexHome = CodexHome()
	}
	prompt = strings.TrimSpace(prompt)

	// Footer first — exclusive to this terminal's scrollback.
	if footer := FindCodexResumeSessionID(scrollback); footer != "" {
		if path, found, err := findCodexTranscriptBySessionID(codexHome, footer); err == nil && found && path != "" {
			// With a known open prompt, reject a stale footer that points at
			// another thread's rollout.
			if prompt == "" || codexRolloutPromptMatches(path, prompt) {
				return footer, true
			}
		} else if prompt == "" {
			// Footer without rollout still usable for resume if codex printed it.
			return footer, true
		}
	}

	if id, _, found, err := scanActiveCodexTranscripts(codexHome, workspace, runStart, prompt); err == nil && found {
		if sid := strings.TrimSpace(id); sid != "" {
			return sid, true
		}
	}
	return "", false
}

// WaitDiscoverCodexSessionID polls until a codex session id is found or budget elapses.
// listenAddr+termSessionID optional: when set, also scrapes scrollback for resume footer.
// Soft: returns "" on timeout (open still succeeds unbound).
func WaitDiscoverCodexSessionID(ctx context.Context, codexHome, workspace string, runStart time.Time, listenAddr, termSessionID, prompt string, budget time.Duration) string {
	if budget <= 0 {
		budget = openCodexBindBudget
	}
	if ctx == nil {
		ctx = context.Background()
	}
	deadline := time.Now().Add(budget)
	for {
		scrollback := ""
		if strings.TrimSpace(listenAddr) != "" && strings.TrimSpace(termSessionID) != "" {
			if snap, err := fetchSnapshotBytes(listenAddr, termSessionID); err == nil {
				scrollback = string(snap)
			}
		}
		if id, ok := DiscoverCodexSessionID(codexHome, workspace, runStart, scrollback, prompt); ok {
			return id
		}
		if time.Now().After(deadline) {
			return ""
		}
		select {
		case <-ctx.Done():
			return ""
		case <-time.After(openCodexBindPoll):
		}
	}
}

func findCodexTranscriptBySessionID(codexHome, sessionID string) (string, bool, error) {
	pattern := filepath.Join(codexHome, "sessions", "*", "*", "*", "rollout-*-"+sessionID+".jsonl")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", false, err
	}
	if len(matches) == 0 {
		return "", false, nil
	}
	return matches[len(matches)-1], true, nil
}

// scanActiveCodexTranscripts finds a cwd-matched rollout after runStart.
// When prompt is non-empty, require one of the first few real user messages
// to match (Codex often stores AGENTS.md as user[1] and the inject as user[2]);
// when multiple candidates remain, fail closed (ok=false) instead of newest-cwd.
// Empty prompt with multiple cwd+time candidates also fails closed.
func scanActiveCodexTranscripts(codexHome, workspace string, runStart time.Time, prompt string) (sessionID, path string, ok bool, err error) {
	pattern := filepath.Join(codexHome, "sessions", "*", "*", "*", "rollout-*.jsonl")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", "", false, err
	}
	absWorkspace, _ := filepath.Abs(canonicalWorkspacePath(workspace))
	prompt = strings.TrimSpace(prompt)
	var candidates []codexTranscriptMeta
	for _, candidate := range matches {
		info, statErr := os.Stat(candidate)
		if statErr != nil || info.IsDir() {
			continue
		}
		meta, metaOK := readCodexSessionMeta(candidate)
		if !metaOK || !codexCwdMatches(meta.Cwd, absWorkspace) {
			continue
		}
		if !sessionNotBefore(runStart, meta.SessionTime) {
			continue
		}
		meta.Path = candidate
		meta.ModTime = info.ModTime()
		if meta.SessionID == "" {
			meta.SessionID = codexSessionIDFromPath(candidate)
		}
		// Prompt gate: skip until one of the first few real user messages
		// matches (poller will retry). Empty prompt skips this filter.
		if prompt != "" && !codexRolloutPromptMatches(candidate, prompt) {
			continue
		}
		candidates = append(candidates, meta)
	}
	if len(candidates) == 0 {
		return "", "", false, nil
	}
	// Multi-candidate fail-closed: never steal newest under concurrent opens.
	if len(candidates) > 1 {
		return "", "", false, nil
	}
	best := candidates[0]
	return best.SessionID, best.Path, true, nil
}

// codexMaxPromptMatchUserMessages is how many non-env user messages to
// compare against the open inject. Codex writes AGENTS.md as user[1] and the
// typed inject as user[2]; 3 covers a preamble + inject without scanning the
// whole transcript.
const codexMaxPromptMatchUserMessages = 3

// codexRolloutPromptMatches reports whether one of the first few real user
// prompts equals want (environment_context-only user blobs are skipped).
func codexRolloutPromptMatches(rolloutPath, want string) bool {
	want = strings.TrimSpace(want)
	if want == "" {
		return true
	}
	for _, got := range codexRealUserPrompts(rolloutPath, codexMaxPromptMatchUserMessages) {
		got = strings.TrimSpace(got)
		if got == want {
			return true
		}
		// Soft: inject may be a prefix/suffix of a longer user blob or vice versa.
		if strings.Contains(got, want) || strings.Contains(want, got) {
			return true
		}
	}
	return false
}

// firstCodexRealUserPrompt returns the first user message text that is not an
// environment_context wrapper.
func firstCodexRealUserPrompt(path string) (string, bool) {
	got := codexRealUserPrompts(path, 1)
	if len(got) == 0 {
		return "", false
	}
	return got[0], true
}

// codexRealUserPrompts returns up to max user message texts from a rollout,
// skipping empty and environment_context-only blobs.
func codexRealUserPrompts(path string, max int) []string {
	if max <= 0 {
		return nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 512*1024)
	var out []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var rec struct {
			Type    string `json:"type"`
			Payload struct {
				Type    string `json:"type"`
				Role    string `json:"role"`
				Message string `json:"message"`
				Content []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"content"`
			} `json:"payload"`
		}
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			continue
		}
		text := ""
		switch rec.Type {
		case "response_item":
			if rec.Payload.Type != "" && rec.Payload.Type != "message" {
				continue
			}
			if strings.TrimSpace(rec.Payload.Role) != "user" {
				continue
			}
			var texts []string
			for _, c := range rec.Payload.Content {
				if t := strings.TrimSpace(c.Text); t != "" {
					texts = append(texts, t)
				}
			}
			text = strings.TrimSpace(strings.Join(texts, "\n"))
		case "event_msg":
			if rec.Payload.Type != "user_message" {
				continue
			}
			text = strings.TrimSpace(rec.Payload.Message)
		default:
			continue
		}
		if text == "" || isCodexEnvironmentContextUserText(text) {
			continue
		}
		out = append(out, text)
		if len(out) >= max {
			break
		}
	}
	return out
}

func isCodexEnvironmentContextUserText(text string) bool {
	return strings.HasPrefix(strings.TrimSpace(text), "<environment_context>")
}

type codexTranscriptMeta struct {
	SessionID   string
	Cwd         string
	Path        string
	SessionTime time.Time
	ModTime     time.Time
}

func readCodexSessionMeta(path string) (codexTranscriptMeta, bool) {
	f, err := os.Open(path)
	if err != nil {
		return codexTranscriptMeta{}, false
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 512*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var rec struct {
			Timestamp string `json:"timestamp"`
			Type      string `json:"type"`
			Payload   struct {
				SessionID string `json:"session_id"`
				ID        string `json:"id"`
				Timestamp string `json:"timestamp"`
				Cwd       string `json:"cwd"`
			} `json:"payload"`
		}
		if err := json.Unmarshal([]byte(line), &rec); err != nil || rec.Type != "session_meta" {
			continue
		}
		if strings.TrimSpace(rec.Payload.Cwd) == "" {
			return codexTranscriptMeta{}, false
		}
		sessionTime, hasSessionTime := parseCodexTimestamp(rec.Payload.Timestamp)
		if !hasSessionTime {
			sessionTime, hasSessionTime = parseCodexTimestamp(rec.Timestamp)
		}
		if !hasSessionTime {
			return codexTranscriptMeta{}, false
		}
		sessionID := strings.TrimSpace(rec.Payload.SessionID)
		if sessionID == "" {
			sessionID = strings.TrimSpace(rec.Payload.ID)
		}
		return codexTranscriptMeta{
			SessionID:   sessionID,
			Cwd:         strings.TrimSpace(rec.Payload.Cwd),
			SessionTime: sessionTime,
		}, true
	}
	return codexTranscriptMeta{}, false
}

func parseCodexTimestamp(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}
	if t, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return t, true
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t, true
	}
	return time.Time{}, false
}

func codexCwdMatches(got, absWorkspace string) bool {
	gotAbs, err := filepath.Abs(canonicalWorkspacePath(got))
	if err != nil {
		return false
	}
	return filepath.Clean(gotAbs) == filepath.Clean(absWorkspace)
}

func codexSessionIDFromPath(path string) string {
	base := strings.TrimSuffix(filepath.Base(path), ".jsonl")
	idx := strings.LastIndex(base, "-")
	if idx < 0 || idx == len(base)-1 {
		return ""
	}
	return base[idx+1:]
}

type codexTranscriptConverter struct {
	seen map[string]struct{}
}

func newCodexTranscriptConverter() *codexTranscriptConverter {
	return &codexTranscriptConverter{seen: make(map[string]struct{})}
}

// TailCodexTranscriptFromOffset tails a codex transcript jsonl from offset.
func TailCodexTranscriptFromOffset(ctx context.Context, path string, startOffset int64, emit func(types.AgentEvent) error) error {
	converter := newCodexTranscriptConverter()
	offset := startOffset
	ticker := time.NewTicker(tailPollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			newOffset, err := readNewCodexTranscript(path, offset, converter, emit)
			if err != nil {
				return err
			}
			offset = newOffset
		}
	}
}

func readNewCodexTranscript(path string, offset int64, converter *codexTranscriptConverter, emit func(types.AgentEvent) error) (int64, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return offset, nil
		}
		return offset, err
	}
	defer f.Close()
	if _, err := f.Seek(offset, 0); err != nil {
		return offset, err
	}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 256*1024), 2*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		for _, ev := range converter.ProcessLine(line) {
			if err := emit(ev); err != nil {
				return offset, err
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return offset, err
	}
	newOffset, err := f.Seek(0, 1)
	if err != nil {
		return offset, err
	}
	return newOffset, nil
}

func (c *codexTranscriptConverter) ProcessLine(line string) []types.AgentEvent {
	var rec struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal([]byte(line), &rec); err != nil {
		return nil
	}
	texts := codexAssistantTexts(rec.Type, rec.Payload)
	events := make([]types.AgentEvent, 0, len(texts))
	for _, text := range texts {
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		if _, ok := c.seen[text]; ok {
			continue
		}
		c.seen[text] = struct{}{}
		events = append(events, types.AgentEvent{
			Type:      types.ActionMessage,
			Role:      "assistant",
			Text:      text,
			Timestamp: time.Now().UnixMilli(),
		})
	}
	return events
}

func codexAssistantTexts(recordType string, payload json.RawMessage) []string {
	switch recordType {
	case "event_msg":
		var p struct {
			Type             string `json:"type"`
			Message          string `json:"message"`
			LastAgentMessage string `json:"last_agent_message"`
		}
		if err := json.Unmarshal(payload, &p); err != nil {
			return nil
		}
		switch p.Type {
		case "agent_message":
			return []string{p.Message}
		case "task_complete":
			return []string{p.LastAgentMessage}
		}
	case "response_item":
		var p struct {
			Type    string `json:"type"`
			Role    string `json:"role"`
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		}
		if err := json.Unmarshal(payload, &p); err != nil {
			return nil
		}
		if p.Type != "message" || p.Role != "assistant" {
			return nil
		}
		var texts []string
		for _, item := range p.Content {
			if item.Type == "output_text" || item.Type == "text" {
				texts = append(texts, item.Text)
			}
		}
		return texts
	}
	return nil
}