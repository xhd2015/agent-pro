package trace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiscoverSourcesFindsHomeAndWorkspaceDotTraceRoots(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	wantRoots := []string{
		filepath.Join(home, ".agent-traces"),
		filepath.Join(home, ".knowledge-hub", "agent-traces"),
		filepath.Join(work, ".agent-traces"),
		filepath.Join(work, ".murphy-codenn", "agent-traces"),
	}
	for _, root := range wantRoots {
		if err := os.MkdirAll(root, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", root, err)
		}
	}

	sources, err := DiscoverSources(home, work)
	if err != nil {
		t.Fatalf("discover sources: %v", err)
	}
	source := NewCombinedSource(sources)
	got := map[string]bool{}
	for _, desc := range source.Describe() {
		got[filepath.Clean(desc)] = true
	}
	for _, want := range wantRoots {
		if !got[filepath.Clean(want)] {
			t.Fatalf("discovered roots = %#v, missing %s", got, want)
		}
	}
}

func TestCombinedSourceListsMultipleRootsWithPrefixedIDs(t *testing.T) {
	firstDataDir := t.TempDir()
	secondDataDir := t.TempDir()
	first, err := StartAgentTraceSession(firstDataDir, AgentTraceMetadata{Command: "first"}, "first prompt")
	if err != nil {
		t.Fatalf("start first session: %v", err)
	}
	first.Finish(nil)
	second, err := StartAgentTraceSession(secondDataDir, AgentTraceMetadata{Command: "second"}, "second prompt")
	if err != nil {
		t.Fatalf("start second session: %v", err)
	}
	second.Finish(nil)

	firstRoot, err := AgentTraceRootForDataDir(firstDataDir)
	if err != nil {
		t.Fatalf("first root: %v", err)
	}
	secondRoot, err := AgentTraceRootForDataDir(secondDataDir)
	if err != nil {
		t.Fatalf("second root: %v", err)
	}
	source := NewCombinedSource([]Source{NewRootSource(firstRoot), NewRootSource(secondRoot)})
	summaries, err := source.List()
	if err != nil {
		t.Fatalf("list combined source: %v", err)
	}
	if len(summaries) != 2 {
		t.Fatalf("summaries = %d, want 2: %#v", len(summaries), summaries)
	}
	for _, summary := range summaries {
		if !strings.Contains(summary.ID, ":") {
			t.Fatalf("summary id %q is not prefixed", summary.ID)
		}
		detail, err := source.Get(summary.ID)
		if err != nil {
			t.Fatalf("get prefixed detail %s: %v", summary.ID, err)
		}
		if detail.Metadata.ID != summary.ID {
			t.Fatalf("detail id = %q, want %q", detail.Metadata.ID, summary.ID)
		}
	}
}

func TestFileSourceParsesJSONLEventsAsReadOnlyTrace(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.jsonl")
	data := strings.Join([]string{
		`{"type":"item.started","item":{"id":"item_1","type":"command_execution","command":"echo ok","status":"in_progress"}}`,
		`{"type":"item.completed","item":{"id":"item_1","type":"command_execution","command":"echo ok","exit_code":0,"aggregated_output":"ok\n"}}`,
		`{"type":"item.completed","item":{"id":"item_2","type":"agent_message","text":"done"}}`,
	}, "\n") + "\n"
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatalf("write events: %v", err)
	}

	source, err := SourceForPath(path)
	if err != nil {
		t.Fatalf("source for file: %v", err)
	}
	summaries, err := source.List()
	if err != nil {
		t.Fatalf("list file source: %v", err)
	}
	if len(summaries) != 1 || summaries[0].LogLineCount != 3 {
		t.Fatalf("summaries = %#v, want one summary with 3 lines", summaries)
	}
	detail, err := source.Get(summaries[0].ID)
	if err != nil {
		t.Fatalf("get file detail: %v", err)
	}
	if len(detail.Messages) != 2 || detail.Messages[1].Content != "done" {
		t.Fatalf("messages = %#v, want parsed shell and assistant output", detail.Messages)
	}
	if _, err := source.Stop(summaries[0].ID); err == nil {
		t.Fatalf("stop file source succeeded, want read-only error")
	}
	if err := source.Delete(summaries[0].ID); err == nil {
		t.Fatalf("delete file source succeeded, want read-only error")
	}
}

func TestSourceForPathDetectsSessionDirWithoutMetadata(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, traceLogFile)
	if err := os.WriteFile(logPath, []byte(`{"type":"result","result":"ok"}`+"\n"), 0o644); err != nil {
		t.Fatalf("write events: %v", err)
	}

	source, err := SourceForPath(dir)
	if err != nil {
		t.Fatalf("source for session dir: %v", err)
	}
	summaries, err := source.List()
	if err != nil {
		t.Fatalf("list session dir source: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("summaries = %d, want 1", len(summaries))
	}
	detail, err := source.Get(summaries[0].ID)
	if err != nil {
		t.Fatalf("get session dir detail: %v", err)
	}
	if len(detail.RawLines) != 1 || len(detail.Messages) != 1 {
		encoded, _ := json.Marshal(detail)
		t.Fatalf("detail did not parse one event: %s", encoded)
	}
}

func TestSourceForPathDiscoversNestedConfigTraceRootsAndLinksChildren(t *testing.T) {
	configHome := t.TempDir()
	murphy, err := StartAgentTraceSession(filepath.Join(configHome, "murphy"), AgentTraceMetadata{
		Command: "murphy",
	}, "murphy prompt")
	if err != nil {
		t.Fatalf("start murphy trace: %v", err)
	}
	murphy.Finish(nil)

	roundDir := filepath.Join(configHome, "codenn", "sessions", "session-1", "rounds", "0001-round")
	codenn, err := StartAgentTraceSession(roundDir, AgentTraceMetadata{
		Command:         "codenn",
		ParentTraceID:   murphy.ID(),
		ParentTraceDir:  murphy.Dir(),
		ParentSessionID: "murphy",
		DelegationID:    "hello-world",
	}, "codenn prompt")
	if err != nil {
		t.Fatalf("start codenn trace: %v", err)
	}
	codenn.Finish(nil)

	source, err := SourceForPath(configHome)
	if err != nil {
		t.Fatalf("source for config home: %v", err)
	}
	descriptions := strings.Join(source.Describe(), "\n")
	if !strings.Contains(descriptions, filepath.Join(configHome, "murphy", "agent-traces")) {
		t.Fatalf("descriptions missing murphy root:\n%s", descriptions)
	}
	if !strings.Contains(descriptions, filepath.Join(roundDir, "agent-traces")) {
		t.Fatalf("descriptions missing codenn root:\n%s", descriptions)
	}

	summaries, err := source.List()
	if err != nil {
		t.Fatalf("list config source: %v", err)
	}
	if len(summaries) != 2 {
		t.Fatalf("summaries = %d, want 2: %#v", len(summaries), summaries)
	}
	var parent AgentTraceSummary
	var child AgentTraceSummary
	for _, summary := range summaries {
		switch summary.Command {
		case "murphy":
			parent = summary
		case "codenn":
			child = summary
		}
	}
	if parent.ID == "" || child.ID == "" {
		t.Fatalf("parent/child summaries not found: %#v", summaries)
	}
	if child.ParentTraceID != parent.ID {
		t.Fatalf("child parent id = %q, want %q", child.ParentTraceID, parent.ID)
	}
	if len(parent.Children) != 1 || parent.Children[0].ID != child.ID || parent.Children[0].DelegationID != "hello-world" {
		t.Fatalf("parent children = %#v, want linked codenn child %q", parent.Children, child.ID)
	}

	detail, err := source.Get(parent.ID)
	if err != nil {
		t.Fatalf("get parent detail: %v", err)
	}
	if len(detail.Metadata.Children) != 1 || detail.Metadata.Children[0].ID != child.ID {
		t.Fatalf("parent detail children = %#v, want child %q", detail.Metadata.Children, child.ID)
	}
	childDetail, err := source.Get(child.ID)
	if err != nil {
		t.Fatalf("get child detail: %v", err)
	}
	if childDetail.Metadata.ParentTraceID != parent.ID {
		t.Fatalf("child detail parent = %q, want %q", childDetail.Metadata.ParentTraceID, parent.ID)
	}
}

func TestFocusSourceIncludesLinkedChildren(t *testing.T) {
	parent := AgentTraceSummary{AgentTraceMetadata: AgentTraceMetadata{
		ID:      "murphy-1",
		Command: "murphy",
		LogPath: filepath.Join("murphy-1", traceLogFile),
	}}
	child := AgentTraceSummary{AgentTraceMetadata: AgentTraceMetadata{
		ID:              "codenn-1",
		Command:         "codenn",
		ParentTraceID:   "murphy-1",
		DelegationID:    "api",
		DelegationLabel: "API",
		LogPath:         filepath.Join("codenn-1", traceLogFile),
	}}
	unrelated := AgentTraceSummary{AgentTraceMetadata: AgentTraceMetadata{
		ID:      "codenn-2",
		Command: "codenn",
		LogPath: filepath.Join("codenn-2", traceLogFile),
	}}

	source := NewFocusSource(staticSource{summaries: []AgentTraceSummary{parent, child, unrelated}}, "murphy", true)
	summaries, err := source.List()
	if err != nil {
		t.Fatalf("list focused source: %v", err)
	}
	if len(summaries) != 2 {
		t.Fatalf("summaries = %#v, want parent and linked child", summaries)
	}
	ids := map[string]bool{}
	for _, summary := range summaries {
		ids[summary.ID] = true
		if summary.ID == "murphy-1" && (len(summary.Children) != 1 || summary.Children[0].DelegationLabel != "API") {
			t.Fatalf("parent children = %#v, want linked labeled child", summary.Children)
		}
	}
	if !ids["murphy-1"] || !ids["codenn-1"] || ids["codenn-2"] {
		t.Fatalf("focused ids = %#v, want murphy-1 and codenn-1 only", ids)
	}
}

func TestRelationshipsGroupSameCodennSessionRoundsAsOneChild(t *testing.T) {
	parent := AgentTraceSummary{AgentTraceMetadata: AgentTraceMetadata{
		ID:      "murphy-1",
		Command: "murphy",
		LogPath: filepath.Join("murphy-1", traceLogFile),
	}}
	firstRound := AgentTraceSummary{AgentTraceMetadata: AgentTraceMetadata{
		ID:              "codenn-round-1",
		Command:         "codenn",
		Status:          "completed",
		ParentTraceID:   "murphy-1",
		DelegationLabel: "api",
		CreatedAt:       "2026-06-05T10:00:00Z",
		LogPath: filepath.Join("config", "codenn", "sessions", "session-1",
			"rounds", "round-1", "agent-traces", "codenn-round-1", traceLogFile),
	}}
	secondRound := AgentTraceSummary{AgentTraceMetadata: AgentTraceMetadata{
		ID:              "codenn-round-2",
		Command:         "codenn",
		Status:          "completed",
		ParentTraceID:   "murphy-1",
		DelegationLabel: "api",
		CreatedAt:       "2026-06-05T10:01:00Z",
		LogPath: filepath.Join("config", "codenn", "sessions", "session-1",
			"rounds", "round-2", "agent-traces", "codenn-round-2", traceLogFile),
	}}

	summaries := withAgentTraceRelationships([]AgentTraceSummary{parent, firstRound, secondRound})
	var linkedParent AgentTraceSummary
	for _, summary := range summaries {
		if summary.ID == "murphy-1" {
			linkedParent = summary
			break
		}
	}
	if len(linkedParent.Children) != 1 {
		t.Fatalf("children = %#v, want one grouped codenn child", linkedParent.Children)
	}
	child := linkedParent.Children[0]
	if child.ID != "codenn-round-2" {
		t.Fatalf("child ID = %q, want latest round codenn-round-2", child.ID)
	}
	if child.DelegationLabel != "api" {
		t.Fatalf("child label = %q, want api", child.DelegationLabel)
	}
}

type staticSource struct {
	summaries []AgentTraceSummary
}

func (s staticSource) List() ([]AgentTraceSummary, error) {
	return s.summaries, nil
}

func (s staticSource) Get(string) (*AgentTraceDetail, error) {
	return nil, nil
}

func (s staticSource) Stop(string) (*AgentTraceDetail, error) {
	return nil, nil
}

func (s staticSource) Delete(string) error {
	return nil
}

func (s staticSource) Describe() []string {
	return nil
}

func TestDynamicRootSourceDiscoversNewRoots(t *testing.T) {
	base := t.TempDir()

	source := NewDynamicRootSource(base)

	summaries, err := source.List()
	if err != nil {
		t.Fatalf("initial list: %v", err)
	}
	if len(summaries) != 0 {
		t.Fatalf("expected 0 summaries initially, got %d", len(summaries))
	}

	roundDir := filepath.Join(base, "session-1", "rounds", "0001-round")
	session, err := StartAgentTraceSession(roundDir, AgentTraceMetadata{
		Command: "codenn",
	}, "prompt")
	if err != nil {
		t.Fatalf("start trace: %v", err)
	}
	session.Finish(nil)

	summaries, err = source.List()
	if err != nil {
		t.Fatalf("list after new trace: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary after adding trace, got %d", len(summaries))
	}
	if summaries[0].Command != "codenn" {
		t.Fatalf("expected codenn command, got %q", summaries[0].Command)
	}

	detail, err := source.Get(summaries[0].ID)
	if err != nil {
		t.Fatalf("get detail: %v", err)
	}
	if detail.Metadata.ID != summaries[0].ID {
		t.Fatalf("detail ID mismatch")
	}
}

func TestDynamicRootSourceKeepsConsistentPrefixes(t *testing.T) {
	base := t.TempDir()

	source := NewDynamicRootSource(base)

	roundDir1 := filepath.Join(base, "session-a", "rounds", "0001-round")
	session1, _ := StartAgentTraceSession(roundDir1, AgentTraceMetadata{Command: "codenn"}, "p1")
	session1.Finish(nil)

	summaries1, _ := source.List()
	if len(summaries1) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries1))
	}
	id1 := summaries1[0].ID
	if !strings.HasPrefix(id1, "s1:") {
		t.Fatalf("expected s1: prefix, got %q", id1)
	}

	roundDir2 := filepath.Join(base, "session-b", "rounds", "0001-round")
	session2, _ := StartAgentTraceSession(roundDir2, AgentTraceMetadata{Command: "codenn"}, "p2")
	session2.Finish(nil)

	summaries2, _ := source.List()
	if len(summaries2) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(summaries2))
	}

	var id1Again, id2 string
	for _, s := range summaries2 {
		switch s.ID {
		case id1:
			id1Again = s.ID
		default:
			id2 = s.ID
		}
	}
	if id1Again != id1 {
		t.Fatalf("first root prefix changed: %q -> %q", id1, id1Again)
	}
	if !strings.HasPrefix(id2, "s2:") {
		t.Fatalf("expected s2: prefix for new root, got %q", id2)
	}
}
