package sessionstate

import (
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/agent/session"
)

func TestReadFromDirMissingMetadata(t *testing.T) {
	sid, _, feat, model, logs := ReadFromDir(t.TempDir(), "fallback")
	if sid != "" || feat != "" || model != "" || logs != nil {
		t.Fatalf("expected empty result, got sid=%q feat=%q model=%q logs=%v", sid, feat, model, logs)
	}
}

func TestReadFromDirUsesFallbackAndReadsEvents(t *testing.T) {
	dir := t.TempDir()
	session.WriteJSON(dir, "metadata.json", Meta{Feature: "Test feature", Model: "gpt-4o"})
	session.AppendLine(dir, "events.jsonl", `{"type":"text","timestamp":1,"sessionID":"sid_1","part":{"text":"hello"}}`)

	sid, _, feat, model, logs := ReadFromDir(dir, "fallback-id")
	if sid != "fallback-id" || feat != "Test feature" || model != "gpt-4o" {
		t.Fatalf("unexpected metadata: sid=%q feat=%q model=%q", sid, feat, model)
	}
	if len(logs) == 0 {
		t.Fatal("expected formatted event logs")
	}
}

func TestReadFromDirNoEvents(t *testing.T) {
	dir := t.TempDir()
	session.WriteJSON(dir, "metadata.json", Meta{SessionID: "sid_2", Feature: "X", Model: "Y"})

	sid, _, feat, model, logs := ReadFromDir(dir, "sid_2")
	if sid != "sid_2" || feat != "X" || model != "Y" {
		t.Fatalf("unexpected metadata: sid=%q feat=%q model=%q", sid, feat, model)
	}
	if logs != nil {
		t.Fatalf("expected nil logs, got %v", logs)
	}
}

func TestResolveEmptyID(t *testing.T) {
	sid, _, _, feat, model, _, err := Resolve("test-agent", "")
	if err != nil {
		t.Fatalf("Resolve empty: %v", err)
	}
	if sid != "" || feat != "" || model != "" {
		t.Fatalf("expected empty values, got sid=%q feat=%q model=%q", sid, feat, model)
	}
}

func TestResolveExistingAndMissing(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("AGENT_PRO_HOME", tmp)

	resumeID := "tcd_abc123"
	dir, err := session.Dir("test-case-design-expert", resumeID)
	if err != nil {
		t.Fatalf("create session dir: %v", err)
	}
	session.WriteJSON(dir, "metadata.json", Meta{SessionID: resumeID, Feature: "Add dark mode", Model: "gpt-4o"})

	sid, _, sdir, feat, model, _, err := Resolve("test-case-design-expert", resumeID)
	if err != nil {
		t.Fatalf("Resolve existing: %v", err)
	}
	if sid != resumeID || sdir != dir || feat != "Add dark mode" || model != "gpt-4o" {
		t.Fatalf("unexpected resolve: sid=%q dir=%q feat=%q model=%q", sid, sdir, feat, model)
	}

	_, _, _, _, _, _, err = Resolve("test-case-design-expert", "missing")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestNewIDPrefixAndUniqueness(t *testing.T) {
	id := NewID("tcd_")
	id2 := NewID("tcd_")
	if !strings.HasPrefix(id, "tcd_") {
		t.Fatalf("id %q should start with tcd_", id)
	}
	if len(id) < 20 {
		t.Fatalf("id too short: %q", id)
	}
	if id == id2 {
		t.Fatal("NewID should produce unique IDs")
	}
}

func TestUpdateOpencodeSessionID(t *testing.T) {
	dir := t.TempDir()
	session.WriteJSON(dir, "metadata.json", Meta{SessionID: "sid", Feature: "F", Model: "M"})
	UpdateOpencodeSessionID(dir, "osid")

	var meta Meta
	if err := session.ReadJSON(dir, "metadata.json", &meta); err != nil {
		t.Fatalf("read metadata: %v", err)
	}
	if meta.OpencodeSessionID != "osid" {
		t.Fatalf("opencode session id = %q, want osid", meta.OpencodeSessionID)
	}
}
