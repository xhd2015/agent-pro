package main

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/openai/openai-go/v3"
	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/agent/llm/llm-mock/mockconfig"
	"github.com/xhd2015/agent-pro/agent/llm/queue"
)

func TestSkillShow(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	oldStdout := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	done := make(chan struct{})
	var buf strings.Builder
	go func() {
		defer close(done)
		io.Copy(&buf, r)
	}()

	if err := handleSkillCommand([]string{"show"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	w.Close()
	<-done

	out := buf.String()
	if !strings.Contains(out, "name: llm-mock") {
		t.Errorf("expected output to contain 'name: llm-mock', got: %s", out)
	}
}

func TestSkillInstallDryRun(t *testing.T) {
	targetDir := t.TempDir()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	oldStdout := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	done := make(chan struct{})
	var buf strings.Builder
	go func() {
		defer close(done)
		io.Copy(&buf, r)
	}()

	if err := handleSkillCommand([]string{"install", "--dry-run", targetDir}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	w.Close()
	<-done

	out := buf.String()
	if !strings.Contains(out, "[dry-run]") {
		t.Errorf("expected output to contain '[dry-run]', got: %s", out)
	}
	if _, statErr := os.Stat(filepath.Join(targetDir, "SKILL.md")); statErr == nil {
		t.Error("expected no SKILL.md after dry run")
	}
}

func TestSkillUnknownSubcommand(t *testing.T) {
	err := handleSkillCommand([]string{"unknown"})
	if err == nil {
		t.Fatal("expected error for unknown subcommand")
	}
	if !strings.Contains(err.Error(), "unknown") {
		t.Errorf("expected error to mention 'unknown', got: %v", err)
	}
}

func TestSkillContentEmbedded(t *testing.T) {
	if skillContent == "" {
		t.Error("skillContent should not be empty")
	}
	if !strings.Contains(skillContent, "name: llm-mock") {
		t.Error("skillContent should contain 'name: llm-mock'")
	}
}

func TestFindMatchEmptyPrefixUsesGeneratorQueue(t *testing.T) {
	h := &mockHandler{
		genQueue: []types.AgentEvent{
			{Type: types.ActionThink, Text: "queued think response"},
			{Type: types.ActionMessage, Text: "queued message response"},
		},
	}
	serve, ok := h.findMatch(userMessages(t, "random-fallback-first-prompt"))
	if !ok || serve == nil {
		t.Fatal("expected generated serve result")
	}
	content := serveMessageContent(serve)
	if !strings.Contains(content, "queued think response") || !strings.Contains(content, "queued message response") {
		t.Fatalf("expected merged think+message content, got %q", content)
	}
	if len(h.genQueue) != 0 {
		t.Fatalf("expected queue drained, left %d", len(h.genQueue))
	}
	if len(serve.agentEvents) != 2 {
		t.Fatalf("expected think+message agent events, got %d", len(serve.agentEvents))
	}
}

func TestFindMatchPrefixThenGeneratorQueue(t *testing.T) {
	content := "from-prefix"
	h := &mockHandler{
		exchanges: []mockconfig.ParsedExchange{
			{
				Exchange: mockconfig.Exchange{
					Request: mockconfig.ExchangeRequest{
						Role:    "user",
						Content: "prefix-only-prompt",
						Index:   -1,
					},
					Response: mockconfig.ExchangeResponse{
						Content:      &content,
						FinishReason: "stop",
					},
				},
			},
		},
		effectiveIndices: []int{0},
		genQueue: []types.AgentEvent{{
			Type: types.ActionMessage,
			Text: "generated after prefix",
		}},
	}

	first, ok := h.findMatch(userMessages(t, "prefix-only-prompt"))
	if !ok || serveMessageContent(first) != "from-prefix" {
		t.Fatalf("first match = %+v, want from-prefix", first)
	}

	second, ok := h.findMatch(userMessages(t, "overflow-to-random-prompt"))
	if !ok || serveMessageContent(second) != "generated after prefix" {
		t.Fatalf("second match = %+v, want generated after prefix", second)
	}
}

func TestFindMatchPrefixNonMatchReturnsNil(t *testing.T) {
	content := "matched"
	h := &mockHandler{
		exchanges: []mockconfig.ParsedExchange{
			{
				Exchange: mockconfig.Exchange{
					Request: mockconfig.ExchangeRequest{
						Role:    "user",
						Content: "only this message",
						Index:   -1,
					},
					Response: mockconfig.ExchangeResponse{
						Content:      &content,
						FinishReason: "stop",
					},
				},
			},
		},
		effectiveIndices: []int{0},
	}

	if serve, ok := h.findMatch(userMessages(t, "completely different message")); ok || serve != nil {
		t.Fatalf("expected no match, got %+v", serve)
	}
}

func TestFindMatchExhaustedGenStreamRenewsPerTurn(t *testing.T) {
	h := &mockHandler{}

	first, ok := h.findMatch(userMessages(t, "Hello"))
	if !ok || serveMessageContent(first) == "" {
		t.Fatalf("first match = %+v, want generated think+message", first)
	}
	if len(first.agentEvents) < 2 {
		t.Fatalf("expected think+message agent events on first serve, got %d", len(first.agentEvents))
	}

	second, ok := h.findMatch(userMessages(t, "Hello"))
	if !ok || serveMessageContent(second) == "" {
		t.Fatalf("second match = %+v, want renewed generated batch", second)
	}

	third, ok := h.findMatch(multiTurnMessages(t,
		"Hello",
		"Here's the result for your request about Hello. I've made the necessary changes.",
		"what's wrong with me?",
	))
	if !ok || serveMessageContent(third) == "" {
		t.Fatalf("third match = %+v, want renewed stream for second user turn", third)
	}
	if serveMessageContent(third) == serveMessageContent(second) {
		t.Fatalf("third match replayed prior merged message %q", serveMessageContent(second))
	}
}

func serveMessageContent(serve *serveResult) string {
	if serve == nil {
		return ""
	}
	if serve.fromLegacy {
		if serve.legacy.Response.Content != nil {
			return *serve.legacy.Response.Content
		}
		return ""
	}
	if serve.batch.Breakpoint.Type != types.ActionMessage {
		return ""
	}
	think := queue.CollapsedThinkText(serve.batch.PrefixThink)
	return think + serve.batch.Breakpoint.Text
}

func multiTurnMessages(t *testing.T, user1, assistant, user2 string) []openai.ChatCompletionMessageParamUnion {
	t.Helper()
	body := []byte(`{"model":"mock-model","messages":[` +
		`{"role":"user","content":"` + user1 + `"},` +
		`{"role":"assistant","content":"` + assistant + `"},` +
		`{"role":"user","content":"` + user2 + `"}` +
		`]}`)
	var req openai.ChatCompletionNewParams
	if err := json.Unmarshal(body, &req); err != nil {
		t.Fatalf("unmarshal messages: %v", err)
	}
	return req.Messages
}

func userMessages(t *testing.T, content string) []openai.ChatCompletionMessageParamUnion {
	t.Helper()
	body := []byte(`{"model":"mock-model","messages":[{"role":"user","content":"` + content + `"}]}`)
	var req openai.ChatCompletionNewParams
	if err := json.Unmarshal(body, &req); err != nil {
		t.Fatalf("unmarshal messages: %v", err)
	}
	return req.Messages
}

func TestRunHelpFlag(t *testing.T) {
	t.Parallel()
	exe := filepath.Join(t.TempDir(), "llm-mock-help-test")
	build := exec.Command("go", "build", "-buildvcs=false", "-o", exe, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}

	cmd := exec.Command(exe, "run", "--help")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run --help: %v\n%s", err, out)
	}
	text := string(out)
	for _, want := range []string{"llm-mock run", "--log-events", "grok", "LLM_MOCK_CONFIG_FILE"} {
		if !strings.Contains(text, want) {
			t.Fatalf("run --help missing %q:\n%s", want, text)
		}
	}
}

func TestMainHelpFlag(t *testing.T) {
	t.Parallel()
	exe := filepath.Join(t.TempDir(), "llm-mock-main-help")
	build := exec.Command("go", "build", "-buildvcs=false", "-o", exe, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}

	cmd := exec.Command(exe, "--help")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("--help: %v\n%s", err, out)
	}
	text := string(out)
	for _, want := range []string{"Usage: llm-mock", "run", "skill", "Server mode"} {
		if !strings.Contains(text, want) {
			t.Fatalf("main --help missing %q:\n%s", want, text)
		}
	}
}
