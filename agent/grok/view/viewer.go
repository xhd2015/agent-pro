// Package view loads Grok CLI sessions, converts updates.jsonl to AgentEvents
// fully in memory, and supports print / follow / read-only web viewing.
package view

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	grok_session "github.com/xhd2015/agent-pro/agent/event/grok_session"
	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/agent/grok/sessions"
	"github.com/xhd2015/dot-pkgs/go-pkgs/logs"
)

// Viewer holds an in-memory conversion of one Grok session.
// It never writes agent-run storage.
type Viewer struct {
	Info        *sessions.SessionInfo
	UpdatesPath string
	GrokHome    string

	mu        sync.RWMutex
	events    []types.AgentEvent
	ndjson    []byte // virtual events.jsonl for byte-offset SSE
	converter *grok_session.Converter
	waiters   []chan struct{}
}

// Open resolves a Grok session and prepares an empty converter.
// Call Bootstrap to load existing updates.jsonl content.
func Open(grokHome, sessionID string) (*Viewer, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, fmt.Errorf("session id is required")
	}
	info, err := sessions.Info(grokHome, sessionID)
	if err != nil {
		return nil, err
	}
	return &Viewer{
		Info:        info,
		UpdatesPath: info.UpdatesPath,
		GrokHome:    grokHome,
		converter:   grok_session.NewConverter(),
	}, nil
}

// Events returns a copy of converted events so far.
func (v *Viewer) Events() []types.AgentEvent {
	v.mu.RLock()
	defer v.mu.RUnlock()
	out := make([]types.AgentEvent, len(v.events))
	copy(out, v.events)
	return out
}

// Offset returns the virtual NDJSON byte length (agent-run events_offset).
func (v *Viewer) Offset() int64 {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return int64(len(v.ndjson))
}

// NDJSONAfter returns virtual events.jsonl bytes after the given offset.
func (v *Viewer) NDJSONAfter(after int64) []byte {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if after < 0 {
		after = 0
	}
	if after >= int64(len(v.ndjson)) {
		return nil
	}
	out := make([]byte, int64(len(v.ndjson))-after)
	copy(out, v.ndjson[after:])
	return out
}

// Subscribe registers for change notifications. Call cancel when done.
func (v *Viewer) Subscribe() (ch <-chan struct{}, cancel func()) {
	c := make(chan struct{}, 1)
	v.mu.Lock()
	v.waiters = append(v.waiters, c)
	v.mu.Unlock()
	var once sync.Once
	return c, func() {
		once.Do(func() {
			v.mu.Lock()
			defer v.mu.Unlock()
			for i, w := range v.waiters {
				if w == c {
					v.waiters = append(v.waiters[:i], v.waiters[i+1:]...)
					break
				}
			}
			close(c)
		})
	}
}

func (v *Viewer) notify() {
	v.mu.RLock()
	waiters := append([]chan struct{}(nil), v.waiters...)
	v.mu.RUnlock()
	for _, w := range waiters {
		select {
		case w <- struct{}{}:
		default:
		}
	}
}

func (v *Viewer) appendEvents(evs []types.AgentEvent) error {
	if len(evs) == 0 {
		return nil
	}
	v.mu.Lock()
	for _, ev := range evs {
		ev = normalizeEvent(ev)
		line, err := json.Marshal(ev)
		if err != nil {
			v.mu.Unlock()
			return err
		}
		v.events = append(v.events, ev)
		v.ndjson = append(v.ndjson, line...)
		v.ndjson = append(v.ndjson, '\n')
	}
	v.mu.Unlock()
	v.notify()
	return nil
}

func normalizeEvent(ev types.AgentEvent) types.AgentEvent {
	if ev.Type == types.ActionMessage && strings.TrimSpace(ev.Role) == "" {
		ev.Role = "assistant"
	}
	if ev.Timestamp == 0 {
		ev.Timestamp = time.Now().UnixMilli()
	}
	ev.Text = strings.TrimRight(ev.Text, "\n")
	ev.Output = strings.TrimRight(ev.Output, "\n")
	return ev
}

func (v *Viewer) processLine(line string) error {
	line = strings.TrimRight(line, "\n")
	if strings.TrimSpace(line) == "" {
		return nil
	}
	if _, ok := grok_session.ParseLine(line); !ok {
		return nil
	}
	return v.appendEvents(v.converter.ProcessLine(line))
}

func (v *Viewer) flushConverter() error {
	return v.appendEvents(v.converter.Flush())
}

// Bootstrap reads existing updates.jsonl from offset 0 into memory.
// Missing file is OK (empty transcript; Follow can wait for creation).
func (v *Viewer) Bootstrap() error {
	path := strings.TrimSpace(v.UpdatesPath)
	if path == "" {
		return nil
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 256*1024), 2*1024*1024)
	for scanner.Scan() {
		if err := v.processLine(scanner.Text()); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return v.flushConverter()
}

// Follow tails updates.jsonl with logs.WatchLine and converts new lines in memory.
// Blocks until ctx is cancelled. Flushes converter on exit.
func (v *Viewer) Follow(ctx context.Context) error {
	path := strings.TrimSpace(v.UpdatesPath)
	if path == "" {
		<-ctx.Done()
		return ctx.Err()
	}

	watchErr := logs.WatchLine(ctx, path, logs.WatchLineOptions{DisableDebounce: true}, func(line string) error {
		return v.processLine(line)
	})

	_ = v.flushConverter()

	if watchErr != nil && watchErr != context.Canceled && ctx.Err() == nil {
		return watchErr
	}
	return ctx.Err()
}

// SessionSummary is the agent-run SessionSummary-compatible meta for the SPA.
func (v *Viewer) SessionSummary() map[string]any {
	info := v.Info
	now := time.Now().UTC().Format(time.RFC3339)
	created := now
	if info != nil && !info.CreatedAt.IsZero() {
		created = info.CreatedAt.UTC().Format(time.RFC3339)
	}
	updated := now
	if info != nil && !info.LastActiveAt.IsZero() {
		updated = info.LastActiveAt.UTC().Format(time.RFC3339)
	}
	workspace := ""
	model := ""
	sessionID := ""
	if info != nil {
		workspace = info.CWD
		model = info.CurrentModelID
		sessionID = info.ID
	}
	return map[string]any{
		"runner":         "grok-tty",
		"session_id":     sessionID,
		"status":         "running",
		"workspace":      workspace,
		"model":          model,
		"created_at":     created,
		"updated_at":     updated,
		"initial_prompt": firstUserText(v.Events()),
	}
}

func firstUserText(events []types.AgentEvent) string {
	for _, ev := range events {
		if ev.Type == types.ActionMessage && ev.Role == "user" {
			return strings.TrimSpace(ev.Text)
		}
	}
	return ""
}

// DetailPayload is agent-run SessionDetail JSON.
func (v *Viewer) DetailPayload() map[string]any {
	events := v.Events()
	if events == nil {
		events = []types.AgentEvent{}
	}
	return map[string]any{
		"session":       v.SessionSummary(),
		"events":        events,
		"events_offset": v.Offset(),
	}
}

// ReadLinesFromOffset walks virtual NDJSON after byte offset and calls onLine
// with each complete line (without trailing newline).
func (v *Viewer) ReadLinesFromOffset(after int64, onLine func(line string) error) error {
	data := v.NDJSONAfter(after)
	if len(data) == 0 {
		return nil
	}
	r := bufio.NewReader(strings.NewReader(string(data)))
	for {
		line, err := r.ReadString('\n')
		if len(line) > 0 {
			trimmed := strings.TrimRight(line, "\n")
			if strings.TrimSpace(trimmed) != "" {
				if err := onLine(trimmed); err != nil {
					return err
				}
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}
