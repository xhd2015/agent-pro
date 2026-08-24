package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const (
	DefaultConfigFile = "config.toml"
)

type Config struct {
	Model                 string                       `toml:"model"`
	ModelReasoningEffort  string                       `toml:"model_reasoning_effort"`
	ModelReasoningSummary string                       `toml:"model_reasoning_summary"`
	ModelVerbosity        string                       `toml:"model_verbosity"`
	DefaultPermissions    string                       `toml:"default_permissions"`
	WebSearch             string                       `toml:"web_search"`
	Personality           string                       `toml:"personality"`
	ApprovalPolicy        string                       `toml:"approval_policy"`
	SandboxMode           string                       `toml:"sandbox_mode"`
	DeveloperInstructions string                       `toml:"developer_instructions"`
	ModelInstructionsFile string                       `toml:"model_instructions_file"`
	ModelProvider         string                       `toml:"model_provider"`
	Notify                []string                     `toml:"notify"`
	ModelProviders        map[string]ModelProvider     `toml:"model_providers"`
	MCPServers            map[string]MCPServer         `toml:"mcp_servers"`
	Projects              map[string]Project           `toml:"projects"`
	Plugins               map[string]Plugin            `toml:"plugins"`
	Marketplaces          map[string]Marketplace       `toml:"marketplaces"`
	Features              Features                     `toml:"features"`
	Permissions           map[string]PermissionProfile `toml:"permissions"`
	Agents                map[string]Agent             `toml:"agents"`
	TUI                   TUISettings                  `toml:"tui"`
	Windows               WindowsSettings              `toml:"windows"`
	Path                  string
}

type ModelProvider struct {
	Name                string            `toml:"name"`
	BaseURL             string            `toml:"base_url"`
	RequiresOpenAIAuth  *bool             `toml:"requires_openai_auth"`
	WireAPI             string            `toml:"wire_api"`
	EnvKey              string            `toml:"env_key"`
	EnvKeyInstructions  string            `toml:"env_key_instructions"`
	HTTPHeaders         map[string]string `toml:"http_headers"`
	EnvHTTPHeaders      map[string]string `toml:"env_http_headers"`
	QueryParams         map[string]string `toml:"query_params"`
	RequestMaxRetries   int               `toml:"request_max_retries"`
	StreamIdleTimeoutMs int               `toml:"stream_idle_timeout_ms"`
	StreamMaxRetries    int               `toml:"stream_max_retries"`
	SupportsWebsockets  bool              `toml:"supports_websockets"`
}

type MCPServer struct {
	Command                 string            `toml:"command"`
	Args                    []string          `toml:"args"`
	URL                     string            `toml:"url"`
	CWD                     string            `toml:"cwd"`
	Env                     map[string]string `toml:"env"`
	EnvVars                 []string          `toml:"env_vars"`
	EnvHTTPHeaders          map[string]string `toml:"env_http_headers"`
	HTTPHeaders             map[string]string `toml:"http_headers"`
	BearerTokenEnvVar       string            `toml:"bearer_token_env_var"`
	Enabled                 *bool             `toml:"enabled"`
	EnabledTools            []string          `toml:"enabled_tools"`
	DisabledTools           []string          `toml:"disabled_tools"`
	DefaultToolsApprovalMode string           `toml:"default_tools_approval_mode"`
	Required                bool              `toml:"required"`
	StartupTimeoutSec       int               `toml:"startup_timeout_sec"`
	ToolTimeoutSec          int               `toml:"tool_timeout_sec"`
	OAuthResource           string            `toml:"oauth_resource"`
	Scopes                  []string          `toml:"scopes"`
	ExperimentalEnvironment string           `toml:"experimental_environment"`
}

// TrustLevel is the per-project trust gate in config.toml [projects."<path>"].
type TrustLevel string

const (
	TrustTrusted   TrustLevel = "trusted"
	TrustUntrusted TrustLevel = "untrusted"
)

// Project is one [projects."<abs-path>"] table.
type Project struct {
	TrustLevel TrustLevel `toml:"trust_level"`
}

type Plugin struct {
	Enabled *bool `toml:"enabled"`
}

type Marketplace struct {
	LastUpdated string `toml:"last_updated"`
	SourceType  string `toml:"source_type"`
	Source      string `toml:"source"`
}

type Features struct {
	Apps                   *bool `toml:"apps"`
	CodexGitCommit         *bool `toml:"codex_git_commit"`
	Hooks                  *bool `toml:"hooks"`
	FastMode               *bool `toml:"fast_mode"`
	Memories               *bool `toml:"memories"`
	MultiAgent             *bool `toml:"multi_agent"`
	Personality            *bool `toml:"personality"`
	ShellSnapshot          *bool `toml:"shell_snapshot"`
	ShellTool              *bool `toml:"shell_tool"`
	UnifiedExec            *bool `toml:"unified_exec"`
	Undo                   *bool `toml:"undo"`
	PreventIdleSleep       *bool `toml:"prevent_idle_sleep"`
	SkillMCPDependencyInstall *bool `toml:"skill_mcp_dependency_install"`
}

type PermissionProfile struct {
	Description string `toml:"description"`
	Extends     string `toml:"extends"`
}

type HookGroup struct {
	Hooks   []HookHandler `toml:"hooks"`
	Matcher string        `toml:"matcher"`
}

type HookHandler struct {
	Type    string `toml:"type"`
	Command string `toml:"command"`
	Timeout int    `toml:"timeout"`
}

type Agent struct {
	ConfigFile        string   `toml:"config_file"`
	Description       string   `toml:"description"`
	NicknameCandidates []string `toml:"nickname_candidates"`
}

type TUISettings struct {
	Keymap TUIKeymap `toml:"keymap"`
}

type TUIKeymap struct {
	Global   map[string]string `toml:"global"`
	Composer map[string]string `toml:"composer"`
}

type WindowsSettings struct {
	Sandbox string `toml:"sandbox"`
}

// ReadFile reads and parses a Codex config.toml from the given path.
func ReadFile(configPath string) (*Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		if os.IsNotExist(err) {
			cfg.Path = configPath
			return &cfg, nil
		}
		return nil, err
	}
	cfg.Path = configPath
	return &cfg, nil
}

// DefaultConfigPath returns the path to the default Codex config file.
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("~", ".codex", DefaultConfigFile)
	}
	return filepath.Join(home, ".codex", DefaultConfigFile)
}

// ReadDefault reads the default config from ~/.codex/config.toml.
func ReadDefault() (*Config, error) {
	return ReadFile(DefaultConfigPath())
}
