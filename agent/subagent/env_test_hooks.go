package subagent

import (
	"os"
	"sync"
)

// getenv resolves env for library paths. When c.EnvLookup is set and returns
// ok=true, that value is used (empty string forces unset for the key without
// falling through to the process environment). When ok=false, falls through
// to os.Getenv. Parallel-safe: each Run passes its own Config/closure.
func getenv(c Config, key string) string {
	if c.EnvLookup != nil {
		if v, ok := c.EnvLookup(key); ok {
			return v
		}
	}
	return os.Getenv(key)
}

func userHomeDir(c Config) (string, error) {
	if c.HomeDir != "" {
		return c.HomeDir, nil
	}
	return os.UserHomeDir()
}

// mockConfigEnvMu serializes FAKE_CODEX_MOCK_CONFIG process mutation for
// child-runner inheritance until agentexec supports per-command ExtraEnv.
var mockConfigEnvMu sync.Mutex

func lockMockConfigEnv(path string) (release func()) {
	mockConfigEnvMu.Lock()
	prev, had := os.LookupEnv("FAKE_CODEX_MOCK_CONFIG")
	_ = os.Setenv("FAKE_CODEX_MOCK_CONFIG", path)
	return func() {
		if had {
			_ = os.Setenv("FAKE_CODEX_MOCK_CONFIG", prev)
		} else {
			_ = os.Unsetenv("FAKE_CODEX_MOCK_CONFIG")
		}
		mockConfigEnvMu.Unlock()
	}
}
