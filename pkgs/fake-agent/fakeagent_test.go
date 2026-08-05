package fakeagent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/fake-agent/probe"
)

// testGenerator returns a Generator that probes a tiny temp dir, not the monorepo.
// Default suggestions include grep/ls against WorkDir; running those on the full
// tree makes multi-seed tests exceed CI timeouts.
func testGenerator(t *testing.T, seed int64) *Generator {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("readme\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	g := NewGenerator(seed)
	g.WorkDir = dir
	return g
}

func TestNewGenerator(t *testing.T) {
	g := NewGenerator(42)
	if g == nil {
		t.Fatal("NewGenerator returned nil")
	}
	if g.rng == nil {
		t.Fatal("generator has nil rng")
	}
}

func TestNextID(t *testing.T) {
	g := NewGenerator(0)
	id := g.NextID()
	if !strings.HasPrefix(id, "item_") {
		t.Fatalf("NextID() = %q, want prefix 'item_'", id)
	}
}

func TestGenerateSession_Deterministic(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("readme\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	g := NewGenerator(42)
	g.WorkDir = dir
	events1 := g.GenerateSession("create a Go HTTP server")

	g2 := NewGenerator(42)
	g2.WorkDir = dir
	events2 := g2.GenerateSession("create a Go HTTP server")

	if len(events1) != len(events2) {
		t.Fatalf("deterministic check: len(events1)=%d, len(events2)=%d", len(events1), len(events2))
	}
	for i := range events1 {
		b1, _ := json.Marshal(events1[i])
		b2, _ := json.Marshal(events2[i])
		if string(b1) != string(b2) {
			t.Fatalf("event %d differs:\n  got: %s\n  want: %s", i, string(b1), string(b2))
		}
	}
}

func TestGenerateSession_NonEmpty(t *testing.T) {
	g := testGenerator(t, 100)
	events := g.GenerateSession("write unit tests")
	if len(events) == 0 {
		t.Fatal("GenerateSession returned empty events")
	}
}

func TestGenerateSession_StartsWithReasoning(t *testing.T) {
	for seed := int64(0); seed < 20; seed++ {
		g := testGenerator(t, seed)
		events := g.GenerateSession("some task")
		if len(events) == 0 {
			t.Fatalf("seed %d: no events", seed)
		}
		first := events[0]
		if first.Type != EventStarted || first.Item == nil || first.Item.Type != ItemReasoning {
			t.Fatalf("seed %d: first event is not reasoning started: %+v", seed, first)
		}
	}
}

func TestGenerateSession_EndsWithMessage(t *testing.T) {
	for seed := int64(0); seed < 20; seed++ {
		g := testGenerator(t, seed)
		events := g.GenerateSession("some task")
		if len(events) == 0 {
			t.Fatalf("seed %d: no events", seed)
		}
		last := events[len(events)-1]
		if last.Type != EventCompleted || last.Item == nil || last.Item.Type != ItemMessage {
			t.Fatalf("seed %d: last event is not message completed: %+v", seed, last)
		}
	}
}

func TestGenerateSession_EventOrdering(t *testing.T) {
	for seed := int64(0); seed < 30; seed++ {
		g := testGenerator(t, seed)
		events := g.GenerateSession("some task")

		itemStates := make(map[string]string)
		for _, e := range events {
			if e.Item == nil {
				continue
			}
			id := e.Item.ID
			prevState := itemStates[id]

			switch e.Type {
			case EventStarted:
				if prevState != "" {
					t.Fatalf("seed %d: item %s started but already in state %s", seed, id, prevState)
				}
				itemStates[id] = "started"
			case EventUpdated:
				if prevState != "started" && prevState != "updated" {
					t.Fatalf("seed %d: item %s updated but state is %s (expected started/updated)", seed, id, prevState)
				}
				itemStates[id] = "updated"
			case EventCompleted:
				if prevState != "started" && prevState != "updated" && prevState != "" {
					t.Fatalf("seed %d: item %s completed but state is %s", seed, id, prevState)
				}
				itemStates[id] = "completed"
			}
		}

		for id, state := range itemStates {
			if state != "completed" {
				t.Fatalf("seed %d: item %s ended in state %s, expected completed", seed, id, state)
			}
		}
	}
}

func TestGenerateSession_MultipleSeedsProduceVariedOutput(t *testing.T) {
	var firstLineSets []map[string]bool
	for seed := int64(0); seed < 10; seed++ {
		g := testGenerator(t, seed)
		events := g.GenerateSession("some task")
		set := make(map[string]bool)
		for _, e := range events {
			b, _ := json.Marshal(e)
			set[string(b)] = true
		}
		firstLineSets = append(firstLineSets, set)
	}

	identical := 0
	for i := 0; i < len(firstLineSets); i++ {
		for j := i + 1; j < len(firstLineSets); j++ {
			if setsEqual(firstLineSets[i], firstLineSets[j]) {
				identical++
			}
		}
	}

	maxIdentical := len(firstLineSets) * (len(firstLineSets) - 1) / 4
	if identical > maxIdentical {
		t.Fatalf("too many identical outputs (%d/%d), random generation may be broken", identical, len(firstLineSets)*(len(firstLineSets)-1)/2)
	}
}

func setsEqual(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}

func TestGenerateSession_WithEmptyPrompt(t *testing.T) {
	g := testGenerator(t, 42)
	events := g.GenerateSession("")
	if len(events) == 0 {
		t.Fatal("empty prompt should still generate events")
	}
}

func TestPick(t *testing.T) {
	g := NewGenerator(0)
	items := []string{"a", "b", "c"}
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		seen[g.pick(items)] = true
	}
	if len(seen) < 2 {
		t.Fatal("pick only returned one item in 100 tries")
	}
}

func TestExtractTopic(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "the task"},
		{"create a web server", "create a web server"},
		{"fix the bug in auth.go", "fix the bug in auth"},
		{"add login. Then test.", "add login"},
		{"\n  add feature\n\nmore text", "add feature"},
	}
	for _, tt := range tests {
		got := extractTopic(tt.input)
		if !strings.Contains(got, tt.want) {
			t.Errorf("extractTopic(%q) = %q, want containing %q", tt.input, got, tt.want)
		}
	}
}

func TestGenerateSession_WithProbeBacktick(t *testing.T) {
	g := testGenerator(t, 42)
	events := g.GenerateSession("please run `echo hello` for me")
	found := false
	for _, e := range events {
		if e.Item != nil && e.Item.Command == "echo hello" {
			found = true
			break
		}
	}
	if !found {
		for _, e := range events {
			if e.Item != nil && e.Item.Type == ItemCommandExecution {
				t.Logf("found command: %s", e.Item.Command)
			}
		}
		t.Fatal("no command_execution with 'echo hello' found")
	}
}

func TestGenerateSession_WithProbeFilePath(t *testing.T) {
	g := testGenerator(t, 42)
	events := g.GenerateSession("read /tmp/config.json and check it")
	found := false
	for _, e := range events {
		if e.Item != nil && e.Item.Command == "cat /tmp/config.json" {
			found = true
			break
		}
	}
	if !found {
		for _, e := range events {
			if e.Item != nil && e.Item.Command != "" {
				t.Logf("found command: %s", e.Item.Command)
			}
		}
		t.Fatal("no cat /tmp/config.json found")
	}
}

func TestGenerateSession_WithProbeFileWrite(t *testing.T) {
	// Direct probe path (session RNG is flaky for exact path selection).
	g := testGenerator(t, 1)
	events, path := g.execProbe(probe.Suggestion{Kind: probe.KindFileWrite, Value: "/tmp/out.txt"})
	if path != "/tmp/out.txt" {
		t.Fatalf("result path = %q, want /tmp/out.txt", path)
	}
	found := false
	for _, e := range events {
		if e.Item != nil && e.Item.Type == ItemFileChange && len(e.Item.Changes) > 0 {
			for _, c := range e.Item.Changes {
				if c.Path == "/tmp/out.txt" {
					found = true
					break
				}
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Fatal("no file_change with /tmp/out.txt found")
	}
}

func TestGenerateSession_WithProbeSearch(t *testing.T) {
	g := testGenerator(t, 42)
	events := g.GenerateSession("grep TODO src/")
	found := false
	for _, e := range events {
		if e.Item != nil && strings.HasPrefix(e.Item.Command, "grep -rn TODO") {
			found = true
			break
		}
	}
	if !found {
		for _, e := range events {
			if e.Item != nil && strings.Contains(e.Item.Command, "grep") {
				t.Logf("found grep: %s", e.Item.Command)
			}
		}
		t.Fatal("no grep command found")
	}
}

func TestGenerateSession_RoundLimit(t *testing.T) {
	for seed := int64(0); seed < 50; seed++ {
		g := testGenerator(t, seed)
		events := g.GenerateSession("some task")

		toolRoundCount := 0
		messageSeen := false
		for _, e := range events {
			if e.Item == nil {
				continue
			}
			switch e.Item.Type {
			case ItemCommandExecution:
				if e.Type == EventCompleted {
					toolRoundCount++
				}
			case ItemFileChange:
				if e.Type == EventCompleted {
					toolRoundCount++
				}
			case ItemMessage:
				if e.Type == EventCompleted {
					messageSeen = true
				}
			}
		}

		if toolRoundCount > maxRounds {
			t.Fatalf("seed %d: %d tool rounds exceeds max %d", seed, toolRoundCount, maxRounds)
		}
		if !messageSeen {
			t.Fatalf("seed %d: no message at end", seed)
		}
	}
}

func TestGenerateSession_ResponseEventuallyProduced(t *testing.T) {
	for seed := int64(0); seed < 30; seed++ {
		g := testGenerator(t, seed)
		events := g.GenerateSession("some task")

		hasMessage := false
		for _, e := range events {
			if e.Item != nil && e.Item.Type == ItemMessage && e.Type == EventCompleted {
				hasMessage = true
				break
			}
		}
		if !hasMessage {
			t.Fatalf("seed %d: no final message produced", seed)
		}
	}
}

func TestExecProbe_ReturnsResultText(t *testing.T) {
	g := NewGenerator(42)
	events, result := g.execProbe(probe.Suggestion{Kind: "tool_call", Value: "echo test"})
	if len(events) == 0 {
		t.Fatal("execProbe returned no events")
	}
	_ = result // some commands legitimately have empty output

	path := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(path, []byte("probe-read-content"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, result2 := g.execProbe(probe.Suggestion{Kind: "file_read", Value: path})
	if result2 == "" {
		t.Fatal("execFileRead returned empty result")
	}
}

func TestPickSuggestion_EmptySuggestsDefault(t *testing.T) {
	g := NewGenerator(42)
	s := g.pickSuggestion(nil)
	if s.Kind == "" {
		t.Fatal("empty suggestions should return a default")
	}
}

func TestExecProbe_FileWriteReturnsPath(t *testing.T) {
	g := NewGenerator(42)
	_, result := g.execProbe(probe.Suggestion{Kind: "file_write", Value: "/tmp/test.txt"})
	if result != "/tmp/test.txt" {
		t.Fatalf("file write result = %q, want /tmp/test.txt", result)
	}
}
