package adhoc_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/agent/codex/adhoc"
	types "github.com/xhd2015/agent-pro/agent/event/types"
	mockwire "github.com/xhd2015/agent-pro/agent/llm/openai"
	"github.com/xhd2015/agent-pro/agent/llm/queue"
)

// This opt-in e2e test uses the installed Codex but ONLY a local mock endpoint.
// It never calls AIS/OpenAI. Temp CODEX_HOME and a fixed small response bound its
// side effects; CommandContext is required for hard cancellation and exact env.
func TestRealCodexAdhocResponsesWireAndToolRoundTrip(t *testing.T) {
	if os.Getenv("AGENT_PRO_TEST_REAL_CODEX") != "1" {
		t.Skip("set AGENT_PRO_TEST_REAL_CODEX=1 for local-only real Codex e2e")
	}
	bin, err := exec.LookPath("codex")
	if err != nil {
		t.Fatal(err)
	}
	home, workspace := t.TempDir(), t.TempDir()
	const secret = "local-fixture-secret-not-a-real-key"
	tokenPath := filepath.Join(t.TempDir(), "token")
	if err := os.WriteFile(tokenPath, []byte(secret), 0600); err != nil {
		t.Fatal(err)
	}
	var wrongCalls atomic.Int32
	wrong := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { wrongCalls.Add(1); http.Error(w, "wrong provider", 500) }))
	defer wrong.Close()
	config := fmt.Sprintf("model_provider = \"global\"\nmodel = \"wrong-model\"\nmodel_reasoning_effort = \"low\"\n[model_providers.global]\nname = \"existing\"\nbase_url = %q\nwire_api = \"responses\"\nrequires_openai_auth = false\n", wrong.URL+"/v1")
	configPath := filepath.Join(home, "config.toml")
	if err := os.WriteFile(configPath, []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	var mu sync.Mutex
	var problems []string
	toolRoundTrip := false
	problem := func(s string) { mu.Lock(); problems = append(problems, s); mu.Unlock() }
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/agent/openai/v1/responses" {
			problem("unexpected request: " + r.Method + " " + r.URL.Path)
			http.Error(w, "unexpected route", 400)
			return
		}
		if r.Header.Get("Authorization") != "Bearer "+secret {
			problem("incorrect bearer credential")
		}
		var body map[string]any
		if json.NewDecoder(r.Body).Decode(&body) != nil {
			problem("invalid request JSON")
			http.Error(w, "bad JSON", 400)
			return
		}
		if body["model"] != "gpt-5.5" {
			problem("display id or wrong model was sent")
		}
		reasoning, _ := body["reasoning"].(map[string]any)
		if reasoning["effort"] != "high" {
			problem(fmt.Sprintf("wrong reasoning: %v", reasoning))
		}
		n := calls.Add(1)
		if n > 3 {
			problem("unexpected repeated inference")
			http.Error(w, "too many requests", 400)
			return
		}
		response := types.AgentEvent{Type: types.ActionMessage, Text: "ADHOC_WIRE_OK"}
		if n == 1 {
			toolName := ""
			if tools, ok := body["tools"].([]any); ok {
				for _, raw := range tools {
					tool, _ := raw.(map[string]any)
					name, _ := tool["name"].(string)
					if name == "exec_command" || name == "shell_command" {
						toolName = name
						break
					}
				}
			}
			if toolName == "" {
				problem(fmt.Sprintf("Codex did not advertise a shell tool; available: %v; request: %v", toolNames(body["tools"]), requestShape(body)))
			} else {
				input := map[string]any{"cmd": "printf AIS_ADHOC_TOOL_OK"}
				if toolName == "shell_command" {
					input = map[string]any{"command": "printf AIS_ADHOC_TOOL_OK"}
				}
				response = types.AgentEvent{Type: types.ActionToolCall, ID: "call_adhoc_fixture", Tool: toolName, ToolInput: input}
			}
		} else {
			if inputs, ok := body["input"].([]any); ok {
				for _, raw := range inputs {
					item, _ := raw.(map[string]any)
					if item["type"] == "function_call_output" {
						value, _ := json.Marshal(item["output"])
						if strings.Contains(string(value), "AIS_ADHOC_TOOL_OK") {
							mu.Lock()
							toolRoundTrip = true
							mu.Unlock()
						}
					}
				}
			}
		}
		if err := mockwire.WriteResponsesStream(w, "gpt-5.5", queue.Batch{Breakpoint: response}); err != nil {
			problem("SSE response failed")
		}
	}))
	defer server.Close()
	c := adhoc.Connection{Version: 1, Provider: "test gateway", BaseURL: server.URL + "/agent/openai/v1", Model: "gpt-5.5", Reasoning: adhoc.ReasoningPolicy{Requested: "high", Effort: "high", Supported: []string{"low", "high"}, Origin: "api"}, Auth: adhoc.Credential{File: tokenPath}}
	argv, err := c.Apply([]string{bin, "exec", "--ignore-rules", "--ephemeral", "--skip-git-repo-check", "--json", "--cd", workspace}, false)
	if err != nil {
		t.Fatal(err)
	}
	argv = append(argv, "Run the shell command requested by the model once, then return its final answer.")
	inherited := []string{}
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "CODEX_HOME=") {
			inherited = append(inherited, e)
		}
	}
	inherited = append(inherited, "CODEX_HOME="+home)
	env, err := c.ChildEnv(inherited)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 70*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, argv[0], argv[1:]...)
	command.Dir = workspace
	command.Env = env
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("real Codex failed: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	mu.Lock()
	defer mu.Unlock()
	if len(problems) > 0 {
		t.Fatalf("wire errors: %v", problems)
	}
	if calls.Load() != 2 || !toolRoundTrip || !strings.Contains(stdout.String(), "ADHOC_WIRE_OK") {
		t.Fatalf("missing SSE/tool roundtrip: calls=%d tool=%v stdout=%s stderr=%s", calls.Load(), toolRoundTrip, stdout.String(), stderr.String())
	}
	if wrongCalls.Load() != 0 {
		t.Fatal("inherited global provider received a request")
	}
	after, err := os.ReadFile(configPath)
	if err != nil || string(after) != config {
		t.Fatal("Codex configuration was rewritten")
	}
	if strings.Contains(stdout.String()+stderr.String(), secret) {
		t.Fatal("credential leaked to output")
	}
	t.Log("e2e: real Codex → local Responses SSE → shell tool → local Responses SSE; global provider/config isolated")
}
