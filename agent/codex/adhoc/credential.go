package adhoc

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Credential is a non-secret reference resolved in the final spawn process.
// PublicToken is ONLY for public proxy-managed dummy values, not API secrets.
type Credential struct {
	EnvKey      string            `json:"env_key,omitempty"`
	File        string            `json:"file,omitempty"`
	JSONPointer string            `json:"json_pointer,omitempty"`
	Guards      map[string]string `json:"guards,omitempty"`
	PublicToken string            `json:"public_token,omitempty"`
	// SHA256 optionally pins unnamed inline credentials against accidental reorder.
	SHA256 string `json:"sha256,omitempty"`
}

var envName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func (c Credential) Empty() bool { return c.EnvKey == "" && c.File == "" && c.PublicToken == "" }
func (c Credential) Validate() error {
	n := 0
	for _, v := range []string{c.EnvKey, c.File, c.PublicToken} {
		if v != "" {
			n++
		}
	}
	if n > 1 {
		return errors.New("credential reference has conflicting sources")
	}
	if c.EnvKey != "" && !envName.MatchString(c.EnvKey) {
		return errors.New("invalid credential environment variable")
	}
	if c.File != "" && !filepath.IsAbs(c.File) {
		return errors.New("credential file reference must be absolute for resume")
	}
	if c.JSONPointer != "" && (c.File == "" || !strings.HasPrefix(c.JSONPointer, "/")) {
		return errors.New("invalid credential JSON reference")
	}
	if len(c.Guards) > 0 && (c.File == "" || c.JSONPointer == "") {
		return errors.New("credential guards require a JSON reference")
	}
	if c.PublicToken != "" && (strings.TrimSpace(c.PublicToken) != c.PublicToken || strings.ContainsAny(c.PublicToken, "\r\n\x00")) {
		return errors.New("invalid public dummy credential")
	}
	return nil
}
func (c Credential) Resolve(environ []string) (string, error) {
	if err := c.Validate(); err != nil {
		return "", err
	}
	token := c.PublicToken
	if c.EnvKey != "" {
		for _, e := range environ {
			if k, v, ok := strings.Cut(e, "="); ok && k == c.EnvKey {
				token = v
			}
		}
		if token == "" {
			return "", errors.New("credential environment variable is missing")
		}
	}
	if c.File != "" {
		f, err := os.Open(c.File)
		if err != nil {
			return "", errors.New("cannot open credential reference")
		}
		defer f.Close()
		body, err := io.ReadAll(io.LimitReader(f, 1<<20+1))
		if err != nil || len(body) > 1<<20 {
			return "", errors.New("cannot read credential reference")
		}
		if c.JSONPointer == "" {
			token = strings.TrimSpace(string(body))
		} else {
			var doc any
			if json.Unmarshal(body, &doc) != nil {
				return "", errors.New("credential source is not valid JSON")
			}
			for p, want := range c.Guards {
				v, ok := pointer(doc, p)
				if !ok || v != want {
					return "", errors.New("credential source identity changed; refusing to retarget the session")
				}
			}
			v, ok := pointer(doc, c.JSONPointer)
			if !ok {
				return "", errors.New("credential JSON reference is missing")
			}
			token = v
		}
		if token == "" {
			return "", errors.New("credential reference is empty")
		}
	}
	if strings.TrimSpace(token) != token || strings.ContainsAny(token, "\r\n\x00") {
		return "", errors.New("credential value is invalid")
	}
	if c.SHA256 != "" {
		h := sha256.Sum256([]byte(token))
		if hex.EncodeToString(h[:]) != c.SHA256 {
			return "", errors.New("credential identity changed; refusing to retarget the session")
		}
	}
	return token, nil
}
func pointer(doc any, path string) (string, bool) {
	if !strings.HasPrefix(path, "/") {
		return "", false
	}
	v := doc
	for _, part := range strings.Split(path[1:], "/") {
		part = strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")
		switch node := v.(type) {
		case map[string]any:
			var ok bool
			v, ok = node[part]
			if !ok {
				return "", false
			}
		case []any:
			i, e := strconv.Atoi(part)
			if e != nil || i < 0 || i >= len(node) {
				return "", false
			}
			v = node[i]
		default:
			return "", false
		}
	}
	s, ok := v.(string)
	return s, ok
}

// ChildEnv removes a stale reserved token and injects this credential only into
// the final Codex child. Never feed the returned slice into persisted RunOptions.Env.
func (c Connection) ChildEnv(inherited []string) ([]string, error) {
	token, err := c.Auth.Resolve(inherited)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(inherited)+1)
	for _, e := range inherited {
		if !strings.HasPrefix(e, TokenEnv+"=") {
			out = append(out, e)
		}
	}
	if !c.Auth.Empty() {
		out = append(out, TokenEnv+"="+token)
	}
	return out, nil
}
