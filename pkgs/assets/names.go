package assets

import "fmt"

// Product and kind constants used for cache paths and release archive names.
const (
	ProductAgentRun = "agent-run"
	ProductAgentPro = "agent-pro"

	KindFrontend = "frontend"
)

// AssetArchiveName returns the release archive basename for a product/kind/version.
// Example: agent-pro_v0.0.70_frontend.tar.gz
func AssetArchiveName(product, version, kind string) string {
	v := NormalizeVersion(version)
	return fmt.Sprintf("%s_%s_%s.tar.gz", product, v, kind)
}

// AssetReleaseURLPath returns the path segment after BaseURL for a release asset.
// Example: /v0.0.70/agent-pro_v0.0.70_frontend.tar.gz
func AssetReleaseURLPath(product, version, kind string) string {
	v := NormalizeVersion(version)
	return fmt.Sprintf("/%s/%s", v, AssetArchiveName(product, version, kind))
}

// AssetReleaseNames returns the frontend archive basenames for agent-run and
// agent-pro for the given version.
func AssetReleaseNames(version string) []string {
	return []string{
		AssetArchiveName(ProductAgentRun, version, KindFrontend),
		AssetArchiveName(ProductAgentPro, version, KindFrontend),
	}
}
