package adhoc

import (
	"errors"
	"strings"
)

// ValidatePersistedEnv prevents callers from accidentally sending the provider
// secret through the legacy persisted Env / env-file transport. File/JSON refs
// or a value inherited by the final wrapper are the supported private sources.
func (c Connection) ValidatePersistedEnv(entries []string) error {
	for _, entry := range entries {
		key, _, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		if key == TokenEnv || (c.Auth.EnvKey != "" && key == c.Auth.EnvKey) {
			return errors.New("provider credential must not use persisted session Env; use a credential reference")
		}
	}
	return nil
}
