package podman

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// FilterEnv returns env with the specified keys removed.
// Each key is matched as a prefix followed by "=".
func FilterEnv(env []string, keys ...string) []string {
	result := make([]string, 0, len(env))
	for _, e := range env {
		skip := false
		for _, key := range keys {
			if strings.HasPrefix(e, key+"=") {
				skip = true
				break
			}
		}
		if !skip {
			result = append(result, e)
		}
	}
	return result
}

// ConfigHash returns the first 8 hex chars of the SHA-256 hash of cfg.
func ConfigHash(cfg string) string {
	h := sha256.Sum256([]byte(cfg))
	return fmt.Sprintf("%x", h[:8])
}
