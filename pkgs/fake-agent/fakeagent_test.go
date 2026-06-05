package fakeagent

import (
	"encoding/json"
	"strings"
	"testing"
)

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
	g := NewGenerator(42)
	events1 := g.GenerateSession("create a Go HTTP server")

	g2 := NewGenerator(42)
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
	g := NewGenerator(100)
	events := g.GenerateSession("write unit tests")
	if len(events) == 0 {
		t.Fatal("GenerateSession returned empty events")
	}
}

func TestGenerateSession_StartsWithReasoning(t *testing.T) {
	for seed := int64(0); seed < 20; seed++ {
		g := NewGenerator(seed)
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
		g := NewGenerator(seed)
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
		g := NewGenerator(seed)
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
		g := NewGenerator(seed)
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
	g := NewGenerator(42)
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
