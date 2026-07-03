package ttyrunner

import (
	"fmt"
	"os"
	"strings"
	"sync"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	agentexec "github.com/xhd2015/agent-pro/agent/exec"
	"github.com/xhd2015/agent-pro/pkgs/groktty"
)

// BuildArgvFunc builds the argv for a TTY runner command inside the PTY.
type BuildArgvFunc func(env *agentexec.Env, settingsPath, agentPath, model, resumeSession string) ([]string, error)

// Provider describes a pluggable TTY runner.
type Provider struct {
	ID             string
	RegistryDir    string
	BannerProvider string
	BannerMarkers  []string
	DisableTail    bool

	BuildArgv          BuildArgvFunc
	StartEventTail     func(ctx TailContext) (runnerSessionID string, err error) // optional stub tail
	DetectScreenStatus func(scrollback []byte) string
	CheckWritable      func(scrollback []byte) WritableStatus
}

var (
	registryMu sync.RWMutex
	providers  []Provider
	byID       = map[string]Provider{}
)

func init() {
	_ = Register(Provider{
		ID:             "grok-tty",
		RegistryDir:    "grok-tty-registry",
		BannerProvider: "grok",
		BannerMarkers:  []string{"GROK_TTY_BANNER"},
		DisableTail:    false,
		BuildArgv:      groktty.BuildGrokCommandArgv,
		DetectScreenStatus: detectGrokScreenStatus,
		CheckWritable:      checkGrokWritable,
	})
	_ = Register(Provider{
		ID:             "codex-tty",
		RegistryDir:    "codex-tty-registry",
		BannerProvider: "codex",
		BannerMarkers:  []string{"CODEX_TTY_BANNER"},
		DisableTail:    true,
		BuildArgv:      groktty.BuildCodexCommandArgv,
		DetectScreenStatus: detectCodexScreenStatus,
		CheckWritable:      checkCodexWritable,
	})
	if os.Getenv("AGENT_RUN_ENABLE_STUB_TTY") == "1" {
		registerStubProvider()
	}
}

// Register adds a TTY runner provider. RegistryDir defaults to ID + "-registry".
func Register(p Provider) error {
	p.ID = strings.TrimSpace(p.ID)
	if p.ID == "" {
		return fmt.Errorf("ttyrunner: provider ID is required")
	}
	if p.RegistryDir == "" {
		p.RegistryDir = p.ID + "-registry"
	}
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, exists := byID[p.ID]; exists {
		return fmt.Errorf("ttyrunner: provider %q already registered", p.ID)
	}
	providers = append(providers, p)
	byID[p.ID] = p
	return nil
}

// Get returns a registered provider by id.
func Get(id string) (Provider, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	p, ok := byID[strings.TrimSpace(id)]
	return p, ok
}

// IDs returns registered provider ids in registration order.
func IDs() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	out := make([]string, len(providers))
	for i, p := range providers {
		out[i] = p.ID
	}
	return out
}

// IsTTYRunner reports whether id is a registered TTY runner.
func IsTTYRunner(id string) bool {
	_, ok := Get(id)
	return ok
}

// sortedProviderIDs returns ids in registration order (internal).
func sortedProviderIDs() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return IDs()
}

// TailContext carries dependencies for provider event tail hooks.
type TailContext struct {
	ScenarioPath string
	Emit         func(types.AgentEvent) error
}

// emitStubLLMEvents is set by stub.go init.
var emitStubLLMEvents func(ctx TailContext) (string, error)

func registerStubProvider() {
	var tail func(TailContext) (string, error)
	if emitStubLLMEvents != nil {
		tail = emitStubLLMEvents
	}
	_ = Register(Provider{
		ID:             "stub-tty",
		RegistryDir:    "stub-tty-registry",
		BannerProvider: "stub",
		BannerMarkers:  []string{"STUB_TTY_BANNER"},
		DisableTail:    true,
		BuildArgv:      BuildStubCommandArgv,
		StartEventTail: tail,
		DetectScreenStatus: detectStubScreenStatus,
		CheckWritable:      checkStubWritable,
	})
}

// Re-register stub when env is set at runtime (doctest harness may set env after init).
func ensureStubRegistered() {
	if os.Getenv("AGENT_RUN_ENABLE_STUB_TTY") != "1" {
		return
	}
	if IsTTYRunner("stub-tty") {
		return
	}
	registerStubProvider()
}

// ProviderListSorted returns providers in registration order.
func ProviderListSorted() []Provider {
	ensureStubRegistered()
	registryMu.RLock()
	defer registryMu.RUnlock()
	out := make([]Provider, len(providers))
	copy(out, providers)
	return out
}



