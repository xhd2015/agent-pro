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
		genQueue: []types.AgentEvent{{
			Type: types.ActionThink,
			Text: "queued think response",
		}},
	}
	ex, _ := h.findMatch(userMessages(t, "random-fallback-first-prompt"))
	if ex == nil {
		t.Fatal("expected generated exchange")
	}
	if ex.Response.Content == nil || *ex.Response.Content != "queued think response" {
		t.Fatalf("expected queued think content, got %+v", ex.Response)
	}
	if len(h.genQueue) != 0 {
		t.Fatalf("expected queue drained, left %d", len(h.genQueue))
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
			Type: types.ActionThink,
			Text: "generated after prefix",
		}},
	}

	first, _ := h.findMatch(userMessages(t, "prefix-only-prompt"))
	if first == nil || first.Response.Content == nil || *first.Response.Content != "from-prefix" {
		t.Fatalf("first match = %+v, want from-prefix", first)
	}

	second, _ := h.findMatch(userMessages(t, "overflow-to-random-prompt"))
	if second == nil || second.Response.Content == nil || *second.Response.Content != "generated after prefix" {
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

	if ex, _ := h.findMatch(userMessages(t, "completely different message")); ex != nil {
		t.Fatalf("expected nil for non-match, got %+v", ex)
	}
}

func TestFindMatchExhaustedGenStreamRenewsPerTurn(t *testing.T) {
	h := &mockHandler{}

	// Turn 1: think then message from one stream (same prompt).
	first, _ := h.findMatch(userMessages(t, "Hello"))
	if first == nil || first.Response.Content == nil || *first.Response.Content == "" {
		t.Fatalf("first match = %+v, want generated think", first)
	}
	second, _ := h.findMatch(userMessages(t, "Hello"))
	if second == nil || second.Response.Content == nil || *second.Response.Content == "" {
		t.Fatalf("second match = %+v, want generated message", second)
	}
	if first.Response.Content == second.Response.Content {
		t.Fatalf("expected distinct think and message content, both %q", *first.Response.Content)
	}

	// Turn 2: stream exhausted; new user prompt must still random-fallback.
	third, _ := h.findMatch(multiTurnMessages(t,
		"Hello",
		"Here's the result for your request about Hello. I've made the necessary changes.",
		"what's wrong with me?",
	))
	if third == nil || third.Response.Content == nil || *third.Response.Content == "" {
		t.Fatalf("third match = %+v, want renewed stream for second user turn", third)
	}
	if *third.Response.Content == *second.Response.Content {
		t.Fatalf("third match replayed turn-1 message %q", *third.Response.Content)
	}
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
	build := exec.Command("go", "build", "-o", exe, ".")
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
	build := exec.Command("go", "build", "-o", exe, ".")
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