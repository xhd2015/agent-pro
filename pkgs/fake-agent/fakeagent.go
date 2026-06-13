package fakeagent

import (
	"fmt"
	"math/rand"
	"strings"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	faketoolexec "github.com/xhd2015/agent-pro/pkgs/fake-agent/fake-tool-exec"
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
	Type    EventType                `json:"type"`
	Item    *EventItem               `json:"item,omitempty"`
	Message string                   `json:"message,omitempty"`
	Text    string                   `json:"text,omitempty"`
	Mock    *faketoolexec.MockConfig `json:"mock,omitempty"`
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

type FileChange = types.FileChange

const (
	maxRounds      = 31
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
		stdout, _, exitCode, _ := faketoolexec.ExecuteBash(s.Value, "", nil)
		return g.buildCmdEvents(id, s.Value, stdout, exitCode)
	case probe.KindFileRead:
		cmd := "cat " + s.Value
		content, err := faketoolexec.ExecuteRead(s.Value)
		if err != nil {
			return g.buildCmdEvents(id, cmd, "", 1)
		}
		return g.buildCmdEvents(id, cmd, content, 0)
	case probe.KindFileWrite:
		faketoolexec.ExecuteWrite(s.Value, "content written by agent")
		return g.buildFileChangeEvents(id, s.Value, "add")
	case probe.KindSearch:
		cmd := "grep -rn " + s.Value + " ."
		output, exitCode, _ := faketoolexec.ExecuteGrep(s.Value, ".")
		return g.buildCmdEvents(id, cmd, output, exitCode)
	default:
		stdout, _, exitCode, _ := faketoolexec.ExecuteBash(s.Value, "", nil)
		return g.buildCmdEvents(id, s.Value, stdout, exitCode)
	}
}

func (g *Generator) buildCmdEvents(id, command, output string, exitCode int) ([]Event, string) {
	ec := exitCode
	started := Event{
		Type: EventStarted,
		Item: &EventItem{ID: id, Type: ItemCommandExecution, Command: command},
	}
	completed := Event{
		Type: EventCompleted,
		Item: &EventItem{
			ID:               id,
			Type:             ItemCommandExecution,
			Command:          command,
			AggregatedOutput: output,
			ExitCode:         &ec,
			Status:           "completed",
		},
	}
	return []Event{started, completed}, output
}

func (g *Generator) buildFileChangeEvents(id, path, kind string) ([]Event, string) {
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
