package assets

import (
	_ "embed"
	"strings"
)

//go:embed VERSION.txt
var versionFile string

// ClientVersion returns the embedded client/module version with a leading "v"
// (e.g. "v0.0.70"). The source of truth is VERSION.txt in this package; keep
// the root module VERSION.txt in sync.
func ClientVersion() string {
	return NormalizeVersion(versionFile)
}

// NormalizeVersion trims whitespace and ensures a leading "v".
// Empty input yields empty string.
func NormalizeVersion(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	if strings.HasPrefix(v, "v") || strings.HasPrefix(v, "V") {
		return "v" + strings.TrimSpace(v[1:])
	}
	return "v" + v
}

// VersionNoPrefix returns the version without a leading "v"/"V".
func VersionNoPrefix(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	if strings.HasPrefix(v, "v") || strings.HasPrefix(v, "V") {
		return strings.TrimSpace(v[1:])
	}
	return v
}
