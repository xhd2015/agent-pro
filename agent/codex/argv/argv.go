// Package argv builds Codex CLI command lines from structured options.
package argv

import (
	"fmt"
	"sort"
	"strings"

	codexcfg "github.com/xhd2015/agent-pro/agent/codex/config"
)

// Options configures a Codex process argv.
// Callers resolve the binary (and optional user flags) then set Bin / UserFlags,
// or set CommandOverride for test/fake TUI hooks.
type Options struct {
	// Bin is argv[0] when CommandOverride is empty.
	Bin string
	// UserFlags are inserted after Bin (settings / runner-spec extras).
	UserFlags []string

	Model           string
	ResumeSession   string
	ReasoningEffort string

	BypassApprovalsAndSandbox bool
	BypassHookTrust           bool
	DisableUpdateCheck        bool
	EmptyMCPServers           bool

	// DisableFeatures appends --disable <name> for each entry (deduped).
	DisableFeatures []string

	// Projects is a typed overlay of config.toml [projects."<path>"].
	// Emitted as -c projects."<path>".trust_level=<level>. Wins over ExtraConfig
	// for the same dotted keys.
	Projects map[string]codexcfg.Project

	// ExtraConfig is an escape hatch for arbitrary -c key=value (dotted paths).
	ExtraConfig map[string]string
	ExtraBool   []string

	// CommandOverride, when set, replaces Bin+UserFlags with shell-split words.
	// Session fields (Model/Resume) are not applied on override; knobs still apply
	// when set on Options (callers choose which knobs for hooks).
	CommandOverride string
}

// Interactive returns agent-run TTY defaults (approvals bypass, no update check,
// hook-trust bypass). Does not empty MCP or disable features.
func Interactive() Options {
	return Options{
		BypassApprovalsAndSandbox: true,
		BypassHookTrust:           true,
		DisableUpdateCheck:        true,
	}
}

// StatusInspect returns the stronger ephemeral /status fetch defaults:
// approvals + hook-trust bypass, no update check, empty mcp_servers, and
// disable plugins / computer_use / in_app_updates / hooks.
func StatusInspect() Options {
	return Options{
		BypassApprovalsAndSandbox: true,
		BypassHookTrust:           true,
		DisableUpdateCheck:        true,
		EmptyMCPServers:           true,
		DisableFeatures: []string{
			"plugins",
			"computer_use",
			"in_app_updates",
			"hooks",
		},
	}
}

// Argv builds the full Codex command line.
func Argv(opts Options) ([]string, error) {
	var args []string
	if override := strings.TrimSpace(opts.CommandOverride); override != "" {
		words, err := ParseShellWords(override)
		if err != nil {
			return nil, err
		}
		args = words
	} else {
		bin := strings.TrimSpace(opts.Bin)
		if bin == "" {
			return nil, fmt.Errorf("codex argv: Bin is required when CommandOverride is empty")
		}
		args = append([]string{bin}, opts.UserFlags...)
		if opts.BypassApprovalsAndSandbox {
			args = EnsureBoolFlag(args, "--dangerously-bypass-approvals-and-sandbox")
		}
		if model := strings.TrimSpace(opts.Model); model != "" && !hasFlagPair(args, "--model") {
			args = append(args, "--model", model)
		}
		if resume := strings.TrimSpace(opts.ResumeSession); resume != "" && !hasCodexResume(args) {
			args = append(args, "resume", resume)
		}
	}

	args = applyKnobs(args, opts)
	return args, nil
}

func applyKnobs(args []string, opts Options) []string {
	// Approvals on override path only (non-override already applied before model/resume).
	if strings.TrimSpace(opts.CommandOverride) != "" && opts.BypassApprovalsAndSandbox {
		args = EnsureBoolFlag(args, "--dangerously-bypass-approvals-and-sandbox")
	}
	// Match historical agent-run order: update-check then hook-trust.
	if opts.DisableUpdateCheck {
		args = EnsureConfigFlag(args, "check_for_update_on_startup", "false")
	}
	if opts.BypassHookTrust {
		args = EnsureBoolFlag(args, "--dangerously-bypass-hook-trust")
	}
	if opts.EmptyMCPServers {
		args = EnsureConfigFlag(args, "mcp_servers", "{}")
	}
	for _, feat := range opts.DisableFeatures {
		feat = strings.TrimSpace(feat)
		if feat == "" {
			continue
		}
		args = EnsureDisableFeature(args, feat)
	}
	if effort := strings.TrimSpace(opts.ReasoningEffort); effort != "" {
		args = EnsureConfigFlag(args, "model_reasoning_effort", effort)
	}
	typedKeys := map[string]struct{}{}
	args = appendProjectsConfig(args, opts.Projects, typedKeys)
	for _, key := range sortedKeys(opts.ExtraConfig) {
		if _, taken := typedKeys[key]; taken {
			continue // typed Projects (etc.) win on conflict
		}
		args = EnsureConfigFlag(args, key, opts.ExtraConfig[key])
	}
	for _, flag := range opts.ExtraBool {
		args = EnsureBoolFlag(args, flag)
	}
	return args
}

// appendProjectsConfig emits -c projects."<path>".trust_level=<level> for each entry.
func appendProjectsConfig(args []string, projects map[string]codexcfg.Project, typedKeys map[string]struct{}) []string {
	if len(projects) == 0 {
		return args
	}
	paths := make([]string, 0, len(projects))
	for p := range projects {
		if strings.TrimSpace(p) == "" {
			continue
		}
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for _, path := range paths {
		proj := projects[path]
		level := strings.TrimSpace(string(proj.TrustLevel))
		if level == "" {
			continue
		}
		key := projectTrustConfigKey(path)
		if typedKeys != nil {
			typedKeys[key] = struct{}{}
		}
		args = EnsureConfigFlag(args, key, level)
	}
	return args
}

// projectTrustConfigKey builds the dotted -c key for a project trust_level.
// Paths are quoted so absolute paths with dots/slashes parse as one map key.
func projectTrustConfigKey(path string) string {
	path = strings.TrimSpace(path)
	escaped := strings.ReplaceAll(path, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	return `projects."` + escaped + `".trust_level`
}

// TrustProject returns a Projects map that marks absPath as trusted.
func TrustProject(absPath string) map[string]codexcfg.Project {
	absPath = strings.TrimSpace(absPath)
	if absPath == "" {
		return nil
	}
	return map[string]codexcfg.Project{
		absPath: {TrustLevel: codexcfg.TrustTrusted},
	}
}

// EnsureTrustedProject appends -c projects."<absPath>".trust_level=trusted when
// absPath is non-empty. Idempotent for the same key.
func EnsureTrustedProject(args []string, absPath string) []string {
	return appendProjectsConfig(args, TrustProject(absPath), nil)
}

func sortedKeys(m map[string]string) []string {
	if len(m) == 0 {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		if strings.TrimSpace(k) == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// EnsureConfigFlag ensures argv has paired "-c" "key=value" (preferred form).
func EnsureConfigFlag(args []string, key, value string) []string {
	key = strings.TrimSpace(key)
	if key == "" {
		return args
	}
	if hasConfigKey(args, key) {
		return args
	}
	out := make([]string, len(args), len(args)+2)
	copy(out, args)
	return append(out, "-c", key+"="+value)
}

// EnsureBoolFlag ensures argv contains a bare boolean flag token exactly once.
func EnsureBoolFlag(args []string, flag string) []string {
	flag = strings.TrimSpace(flag)
	if flag == "" {
		return args
	}
	if hasFlagToken(args, flag) {
		return args
	}
	out := make([]string, len(args), len(args)+1)
	copy(out, args)
	return append(out, flag)
}

// EnsureDisableFeature ensures "--disable" "<name>" appears once.
func EnsureDisableFeature(args []string, name string) []string {
	name = strings.TrimSpace(name)
	if name == "" {
		return args
	}
	if hasDisableFeature(args, name) {
		return args
	}
	out := make([]string, len(args), len(args)+2)
	copy(out, args)
	return append(out, "--disable", name)
}

func hasFlagToken(args []string, flag string) bool {
	for _, a := range args {
		if a == flag {
			return true
		}
	}
	return false
}

func hasFlagPair(args []string, flag string) bool {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == flag {
			return true
		}
	}
	return false
}

func hasCodexResume(args []string) bool {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == "resume" {
			return true
		}
	}
	return false
}

// HasConfigKey reports whether args already contain -c/--config for key
// (either "key=…" or bare "key" as the config value token).
func HasConfigKey(args []string, key string) bool {
	return hasConfigKey(args, key)
}

func hasConfigKey(args []string, key string) bool {
	prefix := key + "="
	for i := 0; i+1 < len(args); i++ {
		if args[i] != "-c" && args[i] != "--config" {
			continue
		}
		v := args[i+1]
		if strings.HasPrefix(v, prefix) || v == key {
			return true
		}
	}
	return false
}

func hasDisableFeature(args []string, name string) bool {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == "--disable" && args[i+1] == name {
			return true
		}
	}
	return false
}

// ParseShellWords splits a shell-style command line into argv words.
func ParseShellWords(line string) ([]string, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, fmt.Errorf("empty command")
	}
	var words []string
	var cur strings.Builder
	var quote rune
	escaped := false
	for _, r := range line {
		if escaped {
			cur.WriteRune(r)
			escaped = false
			continue
		}
		switch {
		case r == '\\' && quote == 0:
			escaped = true
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				cur.WriteRune(r)
			}
		case r == '\'' || r == '"':
			quote = r
		case r == ' ' || r == '\t':
			if cur.Len() > 0 {
				words = append(words, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteRune(r)
		}
	}
	if quote != 0 {
		return nil, fmt.Errorf("unterminated quote in %q", line)
	}
	if cur.Len() > 0 {
		words = append(words, cur.String())
	}
	if len(words) == 0 {
		return nil, fmt.Errorf("empty command")
	}
	return words, nil
}
