package adhoc_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/agent/codex/adhoc"
	types "github.com/xhd2015/agent-pro/agent/event/types"
	mockwire "github.com/xhd2015/agent-pro/agent/llm/openai"
	"github.com/xhd2015/agent-pro/agent/llm/queue"
)

// e2e: GPT-6.1 uses Codex's code-mode tool shape, unlike the conventional shell
// fixture. This verifies its direct Responses/text path without pretending that
// the GPT-5.5 conventional tool fixture proves GPT-6.1 code-mode tool execution.
func TestRealCodexAdhocTextModelAndEffortWire(t *testing.T) {
	if os.Getenv("AGENT_PRO_TEST_REAL_CODEX") != "1" {
		t.Skip("local-only real Codex e2e is opt-in")
	}
	bin, err := exec.LookPath("codex")
	if err != nil {
		t.Fatal(err)
	}
	for _, model := range []string{"gpt-6.1-sol", "upstream/opaque--model"} {
		efforts := []string{"none", "minimal", "low", "medium", "high", "xhigh", "max", "ultra"}
		if model != "gpt-6.1-sol" {
			efforts = []string{"max"}
		}
		for _, effort := range efforts {
			t.Run(model+"/"+effort, func(t *testing.T) {
				var calls atomic.Int32
				var gotEffort atomic.Value
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					calls.Add(1)
					var body map[string]any
					if json.NewDecoder(r.Body).Decode(&body) != nil {
						t.Error("bad request JSON")
						http.Error(w, "bad request", 400)
						return
					}
					if r.Method != "POST" || r.URL.Path != "/v1/responses" || body["model"] != model || r.Header.Get("Authorization") != "Bearer TEST_MANAGED" {
						t.Errorf("incorrect direct connection: %s %s model=%v", r.Method, r.URL.Path, body["model"])
					}
					reasoning, _ := body["reasoning"].(map[string]any)
					value, _ := reasoning["effort"].(string)
					gotEffort.Store(value)
					if err := mockwire.WriteResponsesStream(w, model, queue.Batch{Breakpoint: types.AgentEvent{Type: types.ActionMessage, Text: "GPT61_ADHOC_OK"}}); err != nil {
						t.Error(err)
					}
				}))
				defer server.Close()
				home, workspace := t.TempDir(), t.TempDir()
				config := []byte("model_reasoning_effort = \"low\"\n")
				path := filepath.Join(home, "config.toml")
				if err := os.WriteFile(path, config, 0600); err != nil {
					t.Fatal(err)
				}
				c := adhoc.Connection{Version: 1, Provider: "mock", BaseURL: server.URL + "/v1", Model: model, Reasoning: adhoc.ReasoningPolicy{Requested: effort, Effort: effort, Supported: adhoc.SupportedEfforts(), Origin: "api"}, Auth: adhoc.Credential{PublicToken: "TEST_MANAGED"}}
				if effort == "ultra" {
					if adhoc.SupportsEffort("ultra") {
						t.Fatal("CLI orchestration effort must not be offered as raw API effort ultra")
					}
					c.Reasoning = adhoc.ReasoningPolicy{Requested: "default"}
				}
				argv, err := c.Apply([]string{bin, "exec", "--ignore-rules", "--ephemeral", "--skip-git-repo-check", "--json", "--cd", workspace}, false)
				if err != nil {
					t.Fatal(err)
				}
				if effort == "ultra" {
					argv = append(argv, "-c", `model_reasoning_effort="ultra"`)
				} // probe CLI semantics, not an ad hoc-supported effort
				argv = append(argv, "Return the model's final answer without calling any tools.")
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
				ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
				defer cancel()
				cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
				cmd.Dir = workspace
				cmd.Env = env
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("Codex failed: %v %s", err, out)
				}
				if calls.Load() != 1 || !strings.Contains(string(out), "GPT61_ADHOC_OK") {
					t.Fatalf("missing text SSE result: calls=%d %s", calls.Load(), out)
				}
				actual, _ := gotEffort.Load().(string)
				expected := effort
				if effort == "ultra" {
					expected = "xhigh"
				}
				if actual != expected {
					t.Fatalf("CLI effort %q: expected wire %q, got %q", effort, expected, actual)
				}
				after, err := os.ReadFile(path)
				if err != nil || string(after) != string(config) {
					t.Fatal("config changed")
				}
				t.Log(fmt.Sprintf("e2e: model=%s, CLI effort=%s, wire effort=%s, local SSE only", model, effort, actual))
			})
		}
	}
}
