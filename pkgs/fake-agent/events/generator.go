package events

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	faketoolexec "github.com/xhd2015/agent-pro/pkgs/fake-agent/fake-tool-exec"
	"github.com/xhd2015/agent-pro/pkgs/fake-agent/probe"
)

var probeWriteDir = filepath.Join(os.TempDir(), "fake-agent-probe")

const (
	genMaxRounds      = 31
	genResponseChance = 1.0 / 32.0
)

type Generator struct {
	rng *rand.Rand
}

func NewGenerator(seed int64) *Generator {
	return &Generator{rng: rand.New(rand.NewSource(seed))}
}

func (g *Generator) nextID() string {
	id := g.rng.Int63()
	return fmt.Sprintf("evt_%d", id)
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

func GenerateEvents(seed int64, prompt string) []types.AgentEvent {
	stream := NewEventStream(seed, prompt)
	var events []types.AgentEvent
	for {
		evt, ok := stream.Next()
		if !ok {
			break
		}
		events = append(events, evt)
	}
	return events
}

func (g *Generator) pickSuggestion(suggestions []probe.Suggestion) probe.Suggestion {
	if len(suggestions) == 0 {
		return probe.Suggestion{Kind: "tool_call", Value: "echo no suggestions"}
	}
	return suggestions[g.rng.Intn(len(suggestions))]
}

func (g *Generator) execProbe(s probe.Suggestion) (types.AgentEvent, string) {
	id := g.nextID()
	switch s.Kind {
	case probe.KindToolCall:
		stdout, _, exitCode, _ := faketoolexec.ExecuteBash(s.Value, "", nil)
		ec := exitCode
		return types.AgentEvent{
			ID:       id,
			Type:     types.ActionToolCall,
			Tool:     "bash",
			ToolInput: map[string]any{"command": s.Value},
			Output:   stdout,
			ExitCode: &ec,
		}, stdout
	case probe.KindFileRead:
		content, err := faketoolexec.ExecuteRead(s.Value)
		if err != nil {
			ec := 1
			return types.AgentEvent{
				ID:        id,
				Type:      types.ActionToolCall,
				Tool:      "read",
				ToolInput: map[string]any{"path": s.Value},
				ExitCode:  &ec,
			}, ""
		}
		ec := 0
		return types.AgentEvent{
			ID:        id,
			Type:      types.ActionToolCall,
			Tool:      "read",
			ToolInput: map[string]any{"path": s.Value},
			Output:    content,
			ExitCode:  &ec,
		}, content
	case probe.KindFileWrite:
		os.MkdirAll(probeWriteDir, 0755)
		writePath := filepath.Join(probeWriteDir, filepath.Base(s.Value))
		faketoolexec.ExecuteWrite(writePath, "content written by agent")
		ec := 0
		return types.AgentEvent{
			ID:        id,
			Type:      types.ActionToolCall,
			Tool:      "write",
			ToolInput: map[string]any{"path": s.Value, "content": "content written by agent"},
			ExitCode:  &ec,
		}, s.Value
	case probe.KindSearch:
		output, exitCode, _ := faketoolexec.ExecuteGrep(s.Value, ".")
		ec := exitCode
		return types.AgentEvent{
			ID:        id,
			Type:      types.ActionToolCall,
			Tool:      "grep",
			ToolInput: map[string]any{"pattern": s.Value, "path": "."},
			Output:    output,
			ExitCode:  &ec,
		}, output
	default:
		stdout, _, exitCode, _ := faketoolexec.ExecuteBash(s.Value, "", nil)
		ec := exitCode
		return types.AgentEvent{
			ID:       id,
			Type:     types.ActionToolCall,
			Tool:     "bash",
			ToolInput: map[string]any{"command": s.Value},
			Output:   stdout,
			ExitCode: &ec,
		}, stdout
	}
}

func (g *Generator) generateThink(id, topic string) types.AgentEvent {
	steps := []string{
		fmt.Sprintf("Let me analyze the request regarding %s.", topic),
		fmt.Sprintf("I need to understand the requirements for %s.", topic),
		fmt.Sprintf("Looking at the task: %s. I'll break this down into steps.", topic),
		fmt.Sprintf("First, let me check what's needed for %s.", topic),
		fmt.Sprintf("I'll plan the approach for %s before making changes.", topic),
	}

	fullText := strings.Join(g.shufflePick(steps, 3), "\n")
	return types.AgentEvent{
		ID:   id,
		Type: types.ActionThink,
		Text: fullText,
	}
}

func (g *Generator) generateMessage(id, topic string) types.AgentEvent {
	responses := []string{
		fmt.Sprintf("I've completed the task related to %s. Let me know if you need any adjustments.", topic),
		fmt.Sprintf("Here's the result for your request about %s. I've made the necessary changes.", topic),
		fmt.Sprintf("Done! I've addressed your request regarding %s. The changes should be in place now.", topic),
		fmt.Sprintf("I've analyzed your request about %s and implemented the solution. Please review.", topic),
		fmt.Sprintf("Your request regarding %s has been processed. Everything looks good.", topic),
	}

	text := g.pick(responses)
	return types.AgentEvent{
		ID:   id,
		Type: types.ActionMessage,
		Text: text,
	}
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
	trimmed := strings.TrimSpace(unwrapUserQuery(prompt))
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

// unwrapUserQuery strips grok's <user_query>...</user_query> wrapper so topic
// extraction uses the inner user text instead of the XML tag.
func unwrapUserQuery(prompt string) string {
	trimmed := strings.TrimSpace(prompt)
	const open = "<user_query>"
	const close = "</user_query>"
	if strings.HasPrefix(trimmed, open) && strings.HasSuffix(trimmed, close) {
		return strings.TrimSpace(trimmed[len(open) : len(trimmed)-len(close)])
	}
	return prompt
}
