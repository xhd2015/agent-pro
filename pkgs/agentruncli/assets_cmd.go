package agentruncli

import (
	"context"
	"fmt"
	"strings"
	"time"

	frontend "github.com/xhd2015/agent-pro/frontend-agent-run"
	"github.com/xhd2015/agent-pro/pkgs/assets"
	"github.com/xhd2015/less-gen/flags"
)

const assetsHelp = `
Usage: agent-run assets <command>

Manage downloaded frontend assets for agent-run.

Commands:
  status   show embed + cache completeness (no network)
  ensure   download agent-run frontend into the asset cache if needed

Options:
  -h, --help   show help

Environment:
  AGENT_PRO_ASSET_BASE_URL   base URL for release assets (ensure)
`

func runAssets(args []string) error {
	if len(args) == 0 {
		fmt.Print(strings.TrimPrefix(assetsHelp, "\n"))
		return nil
	}
	cmd := args[0]
	sub := args[1:]
	switch cmd {
	case "-h", "--help":
		fmt.Print(strings.TrimPrefix(assetsHelp, "\n"))
		return nil
	case "status":
		return runAssetsStatus(sub)
	case "ensure":
		return runAssetsEnsure(sub)
	default:
		return fmt.Errorf("unknown assets command: %s (try agent-run assets --help)", cmd)
	}
}

func runAssetsStatus(args []string) error {
	_, err := flags.Help("-h,--help", `
Usage: agent-run assets status

Report embed and on-disk cache completeness for the agent-run frontend.
Does not perform network requests.

Options:
  -h, --help   show help
`).HelpNoExit().Parse(args)
	if err == flags.ErrHelp {
		return nil
	}
	if err != nil {
		return err
	}

	version := assets.ClientVersion()
	product := assets.ProductAgentRun
	kind := assets.KindFrontend

	embedOK := frontend.DistComplete()
	cacheOK := assets.CacheComplete(product, version, kind)
	cacheDir, cacheDirErr := assets.AssetCacheDir(product, version, kind)
	baseURL := assets.ResolveBaseURL(assets.EnsureConfig{})

	fmt.Printf("product:     %s\n", product)
	fmt.Printf("version:     %s\n", version)
	fmt.Printf("kind:        %s\n", kind)
	fmt.Printf("embed:       %s\n", boolComplete(embedOK))
	fmt.Printf("cache:       %s\n", boolComplete(cacheOK))
	if cacheDirErr != nil {
		fmt.Printf("cache_dir:   (error: %v)\n", cacheDirErr)
	} else {
		fmt.Printf("cache_dir:   %s\n", cacheDir)
	}
	if baseURL == "" {
		fmt.Printf("base_url:    (unset; set %s for ensure)\n", assets.EnvAssetBaseURL)
	} else {
		fmt.Printf("base_url:    %s\n", baseURL)
	}
	return nil
}

func runAssetsEnsure(args []string) error {
	_, err := flags.Help("-h,--help", `
Usage: agent-run assets ensure

Download the agent-run frontend SPA into the asset cache when incomplete.
Uses AGENT_PRO_ASSET_BASE_URL (or skips network if cache already complete).

Options:
  -h, --help   show help
`).HelpNoExit().Parse(args)
	if err == flags.ErrHelp {
		return nil
	}
	if err != nil {
		return err
	}

	version := assets.ClientVersion()
	product := assets.ProductAgentRun
	kind := assets.KindFrontend

	if frontend.DistComplete() {
		fmt.Printf("embed is complete; download not required (version %s)\n", version)
		return nil
	}
	if assets.CacheComplete(product, version, kind) {
		dir, _ := assets.AssetCacheDir(product, version, kind)
		fmt.Printf("cache already complete: %s\n", dir)
		return nil
	}

	base := assets.ResolveBaseURL(assets.EnsureConfig{})
	if base == "" {
		return fmt.Errorf("BaseURL is required: set %s or rebuild with a fat frontend embed", assets.EnvAssetBaseURL)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	dir, err := assets.EnsureAsset(ctx, product, version, kind, assets.EnsureConfig{})
	if err != nil {
		return err
	}
	fmt.Printf("ensured: %s\n", dir)
	return nil
}

func boolComplete(ok bool) string {
	if ok {
		return "complete"
	}
	return "incomplete"
}
