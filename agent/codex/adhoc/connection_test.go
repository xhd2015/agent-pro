package adhoc_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/agent/codex/adhoc"
)

func target() adhoc.Connection {
	return adhoc.Connection{Version: 1, Provider: "example", BaseURL: "http://127.0.0.1:9999/agent/openai/v1", Model: "route--gpt-model", Reasoning: adhoc.ReasoningPolicy{Requested: "high", Effort: "high", Supported: []string{"low", "high"}, Origin: "api"}, Auth: adhoc.Credential{EnvKey: "EXAMPLE_PROVIDER_KEY"}}
}

func TestCompileAdhocConnectionPreservesWireIdentityAndIsolatesProvider(t *testing.T) {
	c := target()
	original := []string{"codex", "resume", "sid"}
	before := append([]string(nil), original...)
	argv, err := c.Apply(original, true)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(argv, " ")
	for _, part := range []string{"--no-daemon", "--model route--gpt-model", `wire_api="responses"`, `requires_openai_auth=false`, `supports_websockets=false`, `model_reasoning_effort="high"`, adhoc.TokenEnv, c.BaseURL} {
		if !strings.Contains(joined, part) {
			t.Errorf("missing %q in %s", part, joined)
		}
	}
	if !reflect.DeepEqual(original, before) {
		t.Fatal("mutated caller argv")
	}
	again, err := c.Apply(argv, true)
	if err != nil || !strings.Contains(strings.Join(again, " "), c.ProviderID()) {
		t.Fatalf("revalidation failed: %v", err)
	}
	execArgv, err := c.Apply([]string{"codex", "exec"}, false)
	if err != nil || strings.Contains(strings.Join(execArgv, " "), "--no-daemon") {
		t.Fatalf("exec transport: %v %v", execArgv, err)
	}
	other := c
	other.BaseURL = "http://127.0.0.1:9998/v1"
	if other.ProviderID() == c.ProviderID() {
		t.Fatal("different targets shared provider identity")
	}
	copy := c.Clone()
	copy.Auth.Guards["/name"] = "other"
	copy.Reasoning.Supported[0] = "changed"
	if c.Reasoning.Supported[0] != "low" || len(c.Auth.Guards) != 0 {
		t.Fatal("connection clone aliases caller")
	}
	if levels := adhoc.FilterEfforts([]string{"max", "ultra", "high", "low", "high", "invented"}); !reflect.DeepEqual(levels, []string{"low", "high", "max"}) {
		t.Fatalf("CLI vocabulary: %v", levels)
	}
}

func TestCompileRejectsWrongTargetsAndConflictingOverrides(t *testing.T) {
	for name, modify := range map[string]func(*adhoc.Connection){
		"version":                  func(c *adhoc.Connection) { c.Version = 99 },
		"userinfo":                 func(c *adhoc.Connection) { c.BaseURL = "http://user:secret@example.invalid/v1" },
		"query-secret":             func(c *adhoc.Connection) { c.BaseURL = "http://example.invalid/v1?key=secret" },
		"empty-model":              func(c *adhoc.Connection) { c.Model = "" },
		"unadvertised-max":         func(c *adhoc.Connection) { c.Reasoning.Effort = "max" },
		"ultra-is-not-wire-effort": func(c *adhoc.Connection) { c.Reasoning.Effort = "ultra"; c.Reasoning.Supported = []string{"ultra"} },
		"not-advertised":           func(c *adhoc.Connection) { c.Reasoning.Effort = "xhigh" },
		"unresolved":               func(c *adhoc.Connection) { c.Reasoning.Effort = "" },
	} {
		t.Run(name, func(t *testing.T) {
			c := target()
			modify(&c)
			if _, err := c.Apply([]string{"codex"}, true); err == nil {
				t.Fatal("accepted invalid connection")
			}
		})
	}
	c := target()
	for _, argv := range [][]string{{"codex", "--"}, {"codex", "--model", "other"}, {"codex", "-c", `model_provider="openai"`}, {"codex", "-c", `model_reasoning_effort="low"`}, {"codex", "-c", "model_providers." + c.ProviderID() + ".base_url=wrong"}} {
		if _, err := c.Apply(argv, true); err == nil {
			t.Fatalf("accepted conflict: %v", argv)
		}
	}
}

func TestCredentialIsResolvedOnlyAtFinalSpawnAndCannotUsePersistedEnv(t *testing.T) {
	const secret = "do-not-persist-this-secret"
	path := filepath.Join(t.TempDir(), "source.json")
	body := `{"providers":[{"name":"gateway","baseUrl":"http://example.invalid/v1","token":"` + secret + `"}]}`
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	c := target()
	c.Auth = adhoc.Credential{File: path, JSONPointer: "/providers/0/token", Guards: map[string]string{"/providers/0/name": "gateway", "/providers/0/baseUrl": "http://example.invalid/v1"}}
	argv, err := c.Apply([]string{"codex"}, true)
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(c)
	if strings.Contains(strings.Join(argv, " "), secret) || strings.Contains(string(encoded), secret) {
		t.Fatal("secret escaped into launch description")
	}
	env, err := c.ChildEnv([]string{"PATH=/bin", adhoc.TokenEnv + "=stale"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(env, "\n") != "PATH=/bin\n"+adhoc.TokenEnv+"="+secret {
		t.Fatalf("final env not isolated: %v", err)
	}
	changed := strings.Replace(body, `"gateway"`, `"retargeted"`, 1)
	if err := os.WriteFile(path, []byte(changed), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := c.ChildEnv(nil); err == nil || strings.Contains(err.Error(), secret) {
		t.Fatalf("retarget was not refused safely: %v", err)
	}
	c.Auth = adhoc.Credential{EnvKey: "PRIVATE_KEY"}
	for _, entries := range [][]string{{"PRIVATE_KEY=" + secret}, {adhoc.TokenEnv + "=" + secret}} {
		if err := c.ValidatePersistedEnv(entries); err == nil || strings.Contains(err.Error(), secret) {
			t.Fatal("accepted secret through persisted env")
		}
	}
	if err := c.ValidatePersistedEnv([]string{"TERM=xterm-256color"}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.ChildEnv(nil); err == nil {
		t.Fatal("missing referenced credential did not fail")
	}
	c.Auth = adhoc.Credential{}
	env, err = c.ChildEnv([]string{adhoc.TokenEnv + "=stale", "TERM=xterm"})
	if err != nil || strings.Join(env, ",") != "TERM=xterm" {
		t.Fatal("no-auth inherited stale provider secret")
	}
}
