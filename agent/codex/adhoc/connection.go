// Package adhoc implements process-scoped Codex Responses connections. It never
// creates or rewrites Codex configuration, auth, or model-catalog files.
package adhoc

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

const Version = 1
const TokenEnv = "AGENT_PRO_CODEX_AUTH_TOKEN"

// Connection is safe to persist: Auth contains references or an explicitly
// public dummy credential, never a private bearer token.
type Connection struct {
	Version   int             `json:"version"`
	Provider  string          `json:"provider"`
	BaseURL   string          `json:"base_url"`
	Model     string          `json:"model"`
	Reasoning ReasoningPolicy `json:"reasoning"`
	Auth      Credential      `json:"auth"`
}
type ReasoningPolicy struct {
	Requested string   `json:"requested,omitempty"`
	Effort    string   `json:"effort,omitempty"`
	Supported []string `json:"supported,omitempty"`
	Origin    string   `json:"origin,omitempty"`
}

// SupportedEfforts is the direct Responses wire vocabulary verified with Codex
// 0.166, not an upstream capability claim. max is passed through; ultra is a CLI
// orchestration level that emits xhigh, so it cannot represent API effort ultra.
func SupportedEfforts() []string {
	return []string{"none", "minimal", "low", "medium", "high", "xhigh", "max"}
}
func SupportsEffort(e string) bool {
	for _, v := range SupportedEfforts() {
		if e == v {
			return true
		}
	}
	return false
}
func FilterEfforts(values []string) []string {
	out := []string{}
	for _, e := range SupportedEfforts() {
		for _, v := range values {
			if e == v {
				out = append(out, e)
				break
			}
		}
	}
	return out
}

func (c Connection) Validate() error {
	if c.Version != Version {
		return fmt.Errorf("unsupported Codex connection version %d", c.Version)
	}
	u, err := url.Parse(c.BaseURL)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.RawQuery != "" || u.Fragment != "" || strings.ContainsAny(c.BaseURL, "\r\n\x00") {
		return errors.New("Codex connection needs an http(s) base URL without credentials, query or fragment")
	}
	if c.Model == "" || strings.TrimSpace(c.Model) != c.Model || strings.ContainsAny(c.Model, "\r\n\x00") {
		return errors.New("Codex connection needs an exact nonempty upstream model id")
	}
	if c.Provider == "" || strings.ContainsAny(c.Provider, "\r\n\x00") {
		return errors.New("Codex connection needs a provider identity")
	}
	r := c.Reasoning
	if r.Effort != "" && !SupportsEffort(r.Effort) {
		return fmt.Errorf("reasoning effort %q cannot be passed through unchanged by this Codex adapter; select a supported direct effort", r.Effort)
	}
	if r.Effort != "" && len(r.Supported) > 0 {
		ok := false
		for _, e := range r.Supported {
			ok = ok || e == r.Effort
		}
		if !ok {
			return errors.New("reasoning effort is not supported by the selected model")
		}
	}
	if r.Requested != "" && r.Requested != "default" && r.Effort == "" {
		return errors.New("explicit reasoning choice has not been resolved")
	}
	return c.Auth.Validate()
}
func (c *Connection) Clone() *Connection {
	if c == nil {
		return nil
	}
	out := *c
	out.Reasoning.Supported = append([]string(nil), c.Reasoning.Supported...)
	out.Auth.Guards = map[string]string{}
	for k, v := range c.Auth.Guards {
		out.Auth.Guards[k] = v
	}
	return &out
}

// ProviderID is deterministic across resume and independent of display labels.
func (c Connection) ProviderID() string {
	b, _ := json.Marshal(struct {
		Provider, URL string
		Auth          Credential
	}{c.Provider, c.BaseURL, c.Auth})
	h := sha256.Sum256(b)
	return "agentpro_adhoc_" + hex.EncodeToString(h[:10])
}
func quoted(s string) string { b, _ := json.Marshal(s); return string(b) }

// Apply compiles provider flags without touching files or resolving a secret.
// It rejects conflicting caller flags rather than silently routing elsewhere.
// Call before appending a positional prompt. interactive applies to TUI/resume,
// not to codex exec (which doesn't accept --no-daemon).
func (c Connection) Apply(argv []string, interactive bool) ([]string, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	if len(argv) == 0 {
		return nil, errors.New("Codex argv is empty")
	}
	args := append([]string(nil), argv...)
	for _, arg := range args[1:] {
		if arg == "--" {
			return nil, errors.New("compile the ad hoc connection before appending a prompt separator")
		}
	}
	var model string
	for i := 1; i < len(args); i++ {
		switch {
		case args[i] == "--":
			i = len(args)
		case args[i] == "--model" || args[i] == "-m":
			if i+1 >= len(args) {
				return nil, errors.New("model flag needs a value")
			}
			model = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--model="):
			model = strings.TrimPrefix(args[i], "--model=")
		}
	}
	if model != "" && model != c.Model {
		return nil, errors.New("Codex model flag conflicts with the bound ad hoc connection")
	}
	if model == "" {
		args = append(args, "--model", c.Model)
	}
	id := c.ProviderID()
	provider := `{name=` + quoted(c.Provider) + `,base_url=` + quoted(c.BaseURL) + `,wire_api="responses",requires_openai_auth=false,supports_websockets=false`
	if !c.Auth.Empty() {
		provider += `,env_key=` + quoted(TokenEnv)
	}
	provider += `}`
	owned := map[string]string{"model_provider": quoted(id), "model_providers." + id: provider}
	if c.Reasoning.Effort != "" {
		owned["model_reasoning_effort"] = quoted(c.Reasoning.Effort)
	}
	for i := 1; i < len(args); i++ {
		raw := ""
		if args[i] == "-c" || args[i] == "--config" {
			if i+1 >= len(args) {
				return nil, errors.New("Codex config flag needs a value")
			}
			i++
			raw = args[i]
		} else if strings.HasPrefix(args[i], "--config=") {
			raw = strings.TrimPrefix(args[i], "--config=")
		}
		key, value, ok := strings.Cut(raw, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if want, ok := owned[key]; ok && value != want {
			return nil, fmt.Errorf("Codex config %s conflicts with the bound connection", key)
		}
		if strings.HasPrefix(key, "model_providers."+id+".") {
			return nil, errors.New("Codex provider subkey conflicts with the bound connection")
		}
	}
	// Stable order is useful for diagnostics and wire-test fixtures.
	for _, k := range []string{"model_provider", "model_providers." + id, "model_reasoning_effort"} {
		if v, ok := owned[k]; ok {
			args = append(args, "-c", k+"="+v)
		}
	}
	if interactive {
		found := false
		for _, arg := range args {
			found = found || arg == "--no-daemon"
		}
		if !found {
			args = append([]string{args[0], "--no-daemon"}, args[1:]...)
		}
	}
	return args, nil
}
