// Package commandcode reads Command Code (cmd) account data over the public
// HTTP API, so account commands work without the Node CLI installed.
package commandcode

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// DefaultBaseURL is the production Command Code API endpoint.
	DefaultBaseURL = "https://api.commandcode.ai"
	// DefaultStudioHost serves the account usage page opened by --open.
	DefaultStudioHost = "https://commandcode.ai"
	// APIKeyEnvVar overrides the apiKey stored in auth.json.
	APIKeyEnvVar = "COMMAND_CODE_API_KEY"
	// AuthFileName is the credentials file the cmd CLI writes.
	AuthFileName = "auth.json"
	// UserAgent is sent on every request. Cloudflare answers key-only
	// requests that carry no User-Agent with 403 error code: 1010.
	UserAgent = "cli"

	sandboxEnvVar = "COMMANDCODE_SANDBOX"
	apiURLEnvVar  = "COMMANDCODE_API_URL"
)

// Auth holds the credentials read for one Command Code home.
type Auth struct {
	APIKey          string `json:"apiKey"`
	UserID          string `json:"userId"`
	UserName        string `json:"userName"`
	KeyName         string `json:"keyName"`
	AuthenticatedAt string `json:"authenticatedAt"`
	// Source is the origin of APIKey: "env" or the auth.json path.
	Source string `json:"-"`
}

// DefaultHome returns the Command Code config directory (~/.commandcode).
func DefaultHome() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".commandcode")
}

// ResolveHome expands ~ and makes home absolute. An empty home means
// DefaultHome.
func ResolveHome(home string) (string, error) {
	home = strings.TrimSpace(home)
	if home == "" {
		home = DefaultHome()
	}
	if home == "" {
		return "", fmt.Errorf("cannot determine the user home directory; pass --home")
	}
	if home == "~" || strings.HasPrefix(home, "~/") {
		userHome, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("expand %q: %w", home, err)
		}
		home = filepath.Join(userHome, strings.TrimPrefix(strings.TrimPrefix(home, "~"), "/"))
	}
	abs, err := filepath.Abs(home)
	if err != nil {
		return "", fmt.Errorf("resolve home %q: %w", home, err)
	}
	return abs, nil
}

// AuthPath returns the auth.json path inside home.
func AuthPath(home string) string {
	return filepath.Join(home, AuthFileName)
}

// ReadAuth reads credentials for home, preferring $COMMAND_CODE_API_KEY.
func ReadAuth(home string) (*Auth, error) {
	if key := strings.TrimSpace(os.Getenv(APIKeyEnvVar)); key != "" {
		return &Auth{APIKey: key, Source: "env " + APIKeyEnvVar}, nil
	}
	path := AuthPath(home)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("not authenticated: no credentials at %s; run `cmd login` first, or set %s", path, APIKeyEnvVar)
		}
		return nil, fmt.Errorf("read credentials %s: %w", path, err)
	}
	var auth Auth
	if err := json.Unmarshal(data, &auth); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	auth.APIKey = strings.TrimSpace(auth.APIKey)
	if auth.APIKey == "" {
		return nil, fmt.Errorf("not authenticated: %s has no apiKey; run `cmd login` first", path)
	}
	auth.Source = path
	return &auth, nil
}

// ResolveBaseURL returns apiURL when set, else the sandbox override the cmd
// CLI honors, else DefaultBaseURL.
func ResolveBaseURL(apiURL string) string {
	if url := strings.TrimSpace(apiURL); url != "" {
		return strings.TrimRight(url, "/")
	}
	if strings.EqualFold(strings.TrimSpace(os.Getenv(sandboxEnvVar)), "true") {
		if url := strings.TrimSpace(os.Getenv(apiURLEnvVar)); url != "" {
			return strings.TrimRight(url, "/")
		}
	}
	return DefaultBaseURL
}
