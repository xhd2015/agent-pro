package fakeagent

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/fake-agent/probe"
)

type EventType string

const (
	EventStarted   EventType = "item.started"
	EventUpdated   EventType = "item.updated"
	EventCompleted EventType = "item.completed"
	EventError     EventType = "error"
)

type ItemType string

const (
	ItemReasoning        ItemType = "reasoning"
	ItemCommandExecution ItemType = "command_execution"
	ItemFileChange       ItemType = "file_change"
	ItemMessage          ItemType = "message"
)

type Event struct {
	Type    EventType  `json:"type"`
	Item    *EventItem `json:"item,omitempty"`
	Message string     `json:"message,omitempty"`
	Text    string     `json:"text,omitempty"`
}

type EventItem struct {
	ID               string       `json:"id"`
	Type             ItemType     `json:"type"`
	Text             string       `json:"text,omitempty"`
	Content          []ItemPart   `json:"content,omitempty"`
	Command          string       `json:"command,omitempty"`
	AggregatedOutput string       `json:"aggregated_output,omitempty"`
	ExitCode         *int         `json:"exit_code,omitempty"`
	Status           string       `json:"status,omitempty"`
	Changes          []FileChange `json:"changes,omitempty"`
}

type ItemPart struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type FileChange struct {
	Path string `json:"path"`
	Kind string `json:"kind"`
}

const (
	maxRounds     = 31
	responseChance = 1.0 / 32.0
)

type Generator struct {
	rng *rand.Rand
}

func NewGenerator(seed int64) *Generator {
	return &Generator{rng: rand.New(rand.NewSource(seed))}
}

func (g *Generator) NextID() string {
	id := g.rng.Int63()
	return fmt.Sprintf("item_%d", id)
}

func (g *Generator) pick(items []string) string {
	if len(items) == 0 {
		return ""
	}
	return items[g.rng.Intn(len(items))]
}

func (g *Generator) chance(p float64) bool {
	return g.rng.Float64() < p
}

func (g *Generator) GenerateSession(prompt string) []Event {
	var events []Event

	topic := extractTopic(prompt)

	reasoningID := g.NextID()
	reasoningEvents := g.generateReasoning(reasoningID, topic)
	events = append(events, reasoningEvents...)

	probeList := probe.Scan(prompt)
	probeList = probe.Merge(probeList, probe.DefaultSuggestions())

	for round := 0; round < maxRounds; round++ {
		if g.chance(responseChance) {
			break
		}

		suggestion := g.pickSuggestion(probeList)
		toolEvents, result := g.execProbe(suggestion)
		events = append(events, toolEvents...)

		if result != "" {
			newProbes := probe.Scan(result)
			probeList = probe.Merge(probeList, newProbes)
		}
	}

	msgID := g.NextID()
	msgEvents := g.generateMessage(msgID, topic)
	events = append(events, msgEvents...)

	return events
}

func (g *Generator) pickSuggestion(suggestions []probe.Suggestion) probe.Suggestion {
	if len(suggestions) == 0 {
		return probe.Suggestion{Kind: "tool_call", Value: "echo no suggestions"}
	}
	return suggestions[g.rng.Intn(len(suggestions))]
}

func (g *Generator) execProbe(s probe.Suggestion) ([]Event, string) {
	id := g.NextID()
	switch s.Kind {
	case probe.KindToolCall:
		return g.execToolCall(id, s.Value)
	case probe.KindFileRead:
		return g.execFileRead(id, s.Value)
	case probe.KindFileWrite:
		return g.execFileWrite(id, s.Value)
	case probe.KindSearch:
		return g.execSearch(id, s.Value)
	default:
		return g.execRandomTool(id)
	}
}

func (g *Generator) execToolCall(id, cmd string) ([]Event, string) {
	stdout := fakes[g.rng.Intn(len(fakes))]
	started := Event{
		Type: EventStarted,
		Item: &EventItem{ID: id, Type: ItemCommandExecution, Command: cmd},
	}
	exitCode := 0
	completed := Event{
		Type: EventCompleted,
		Item: &EventItem{
			ID:               id,
			Type:             ItemCommandExecution,
			Command:          cmd,
			AggregatedOutput: stdout.text,
			ExitCode:         &exitCode,
			Status:           "completed",
		},
	}
	return []Event{started, completed}, stdout.text
}

func (g *Generator) execFileRead(id, path string) ([]Event, string) {
	cmd := "cat " + path
	stdout := fakeReadContent[g.rng.Intn(len(fakeReadContent))]
	started := Event{
		Type: EventStarted,
		Item: &EventItem{ID: id, Type: ItemCommandExecution, Command: cmd},
	}
	exitCode := 0
	completed := Event{
		Type: EventCompleted,
		Item: &EventItem{
			ID:               id,
			Type:             ItemCommandExecution,
			Command:          cmd,
			AggregatedOutput: stdout.text,
			ExitCode:         &exitCode,
			Status:           "completed",
		},
	}
	return []Event{started, completed}, stdout.text
}

func (g *Generator) execFileWrite(id, path string) ([]Event, string) {
	kind := "add"
	if g.rng.Intn(2) == 0 {
		kind = "modify"
	}
	started := Event{
		Type: EventStarted,
		Item: &EventItem{ID: id, Type: ItemFileChange},
	}
	completed := Event{
		Type: EventCompleted,
		Item: &EventItem{
			ID:      id,
			Type:    ItemFileChange,
			Status:  "completed",
			Changes: []FileChange{{Path: path, Kind: kind}},
		},
	}
	return []Event{started, completed}, path
}

func (g *Generator) execSearch(id, query string) ([]Event, string) {
	cmd := "grep -rn " + query + " ."
	stdout := fakeSearchResults[g.rng.Intn(len(fakeSearchResults))]
	started := Event{
		Type: EventStarted,
		Item: &EventItem{ID: id, Type: ItemCommandExecution, Command: cmd},
	}
	exitCode := 0
	completed := Event{
		Type: EventCompleted,
		Item: &EventItem{
			ID:               id,
			Type:             ItemCommandExecution,
			Command:          cmd,
			AggregatedOutput: stdout,
			ExitCode:         &exitCode,
			Status:           "completed",
		},
	}
	return []Event{started, completed}, stdout
}

func (g *Generator) execRandomTool(id string) ([]Event, string) {
	defaultCommands := []string{
		"ls -la",
		"git status",
		"git diff",
		"cat README.md",
		"find . -name \"*.go\"",
		"go test ./...",
		"go build ./...",
	}
	cmd := defaultCommands[g.rng.Intn(len(defaultCommands))]
	return g.execToolCall(id, cmd)
}

func (g *Generator) generateReasoning(id, topic string) []Event {
	steps := []string{
		fmt.Sprintf("Let me analyze the request regarding %s.", topic),
		fmt.Sprintf("I need to understand the requirements for %s.", topic),
		fmt.Sprintf("Looking at the task: %s. I'll break this down into steps.", topic),
		fmt.Sprintf("First, let me check what's needed for %s.", topic),
		fmt.Sprintf("I'll plan the approach for %s before making changes.", topic),
	}

	started := Event{
		Type: EventStarted,
		Item: &EventItem{
			ID:   id,
			Type: ItemReasoning,
		},
	}

	var updates []Event
	for i := 0; i < g.rng.Intn(3)+1; i++ {
		updates = append(updates, Event{
			Type: EventUpdated,
			Item: &EventItem{
				ID:   id,
				Type: ItemReasoning,
				Text: g.pick(steps),
			},
		})
	}

	fullText := strings.Join(g.shufflePick(steps, 3), "\n")
	completed := Event{
		Type: EventCompleted,
		Item: &EventItem{
			ID:     id,
			Type:   ItemReasoning,
			Text:   fullText,
			Status: "completed",
		},
	}

	var events []Event
	events = append(events, started)
	events = append(events, updates...)
	events = append(events, completed)
	return events
}

func (g *Generator) generateMessage(id, topic string) []Event {
	responses := []string{
		fmt.Sprintf("I've completed the task related to %s. Let me know if you need any adjustments.", topic),
		fmt.Sprintf("Here's the result for your request about %s. I've made the necessary changes.", topic),
		fmt.Sprintf("Done! I've addressed your request regarding %s. The changes should be in place now.", topic),
		fmt.Sprintf("I've analyzed your request about %s and implemented the solution. Please review.", topic),
		fmt.Sprintf("Your request regarding %s has been processed. Everything looks good.", topic),
	}

	text := g.pick(responses)

	completed := Event{
		Type: EventCompleted,
		Item: &EventItem{
			ID:      id,
			Type:    ItemMessage,
			Text:    text,
			Content: []ItemPart{{Type: "output_text", Text: text}},
			Status:  "completed",
		},
	}

	return []Event{completed}
}

func (g *Generator) shufflePick(items []string, n int) []string {
	if n > len(items) {
		n = len(items)
	}
	perm := g.rng.Perm(len(items))
	result := make([]string, n)
	for i := 0; i < n; i++ {
		result[i] = items[perm[i]]
	}
	return result
}

type fakeCommand struct {
	cmd  string
	text string
}

var fakes = []fakeCommand{
	{"ls -la", "src/\n  main.go\n  utils.go\nREADME.md\n"},
	{"git status", "On branch main\nnothing to commit, working tree clean\n"},
	{"git diff", ""},
	{"cat README.md", "# Project\n\nThis is a sample project.\nSee docs/ for /tmp/details.\n"},
	{"find . -name \"*.go\"", "./cmd/main.go\n./pkgs/foo/foo.go\n"},
	{"go test ./...", "ok \t github.com/xhd2015/agent-pro/... \t 0.123s\n"},
	{"go build ./...", ""},
}

var fakeReadContent = []fakeCommand{
	{"cat config.json", "{\n  \"version\": \"1.0\",\n  \"include\": [\"/tmp/shared\", \"/tmp/plugins\"]\n}\n"},
	{"cat Makefile", "build:\n\tgo build -o /tmp/output ./cmd/...\n\ntest:\n\tgo test ./...\n"},
	{"cat .env", "DATABASE_URL=postgres://localhost/db\nAPI_KEY=sk-xxx\n"},
	{"cat TODO.md", "# TODO\n\n- fix bug in /tmp/auth.go\n- add /tmp/feature.go\n- search for deprecated\n"},
}

var fakeSearchResults = []string{
	"src/main.go:10: import \"github.com/xhd2015/agent-pro\"\nsrc/main.go:42: TODO: refactor /tmp/legacy.go\n",
	"README.md:5: See /tmp/docs/guide.md for setup.\nREADME.md:20: Run `go test ./tmp/...`\n",
	"config.toml:3: path = \"/tmp/data\"\nconfig.toml:8: include = \"/tmp/extra\"\n",
	"",
}

func extractTopic(prompt string) string {
	trimmed := strings.TrimSpace(prompt)
	if trimmed == "" {
		return "the task"
	}
	stopWords := []string{".", "\n", "?", "!"}
	end := len(trimmed)
	for _, w := range stopWords {
		if idx := strings.Index(trimmed, w); idx >= 0 && idx < end {
			end = idx
		}
	}
	topic := strings.TrimSpace(trimmed[:end])
	if len(topic) > 50 {
		topic = topic[:50] + "..."
	}
	return topic
}
