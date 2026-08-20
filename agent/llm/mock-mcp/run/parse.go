package run

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	lessflags "github.com/xhd2015/less-flags"
)

const (
	EnvName     = "MOCK_MCP_NAME"
	EnvDelay    = "MOCK_MCP_DELAY"
	EnvDelayMin = "MOCK_MCP_DELAY_MIN"
	EnvDelayMax = "MOCK_MCP_DELAY_MAX"
	EnvHang     = "MOCK_MCP_HANG"
	EnvFail     = "MOCK_MCP_FAIL"
	EnvDebug    = "MOCK_MCP_DEBUG"
)

const help = `
Usage: mock-mcp [options]

Stdio MCP mock. Sleeps before the initialize result so Codex chrome
stays on "Starting MCP servers".

Options:
  --name NAME              serverInfo.name (env: MOCK_MCP_NAME)
  --delay DURATION         fixed delay (env: MOCK_MCP_DELAY)
  --delay-min DURATION     random lower bound (env: MOCK_MCP_DELAY_MIN)
  --delay-max DURATION     random upper bound (env: MOCK_MCP_DELAY_MAX)
  --hang                   never answer initialize (env: MOCK_MCP_HANG)
  --fail                   fail after delay (env: MOCK_MCP_FAIL)
  -h, --help               show this help
`

const defaultName = "mock-mcp"

// Config is the resolved mock-mcp process config (CLI overrides env).
type Config struct {
	Name     string
	Delay    *time.Duration
	DelayMin *time.Duration
	DelayMax *time.Duration
	Hang     bool
	Fail     bool
}

func Main(args []string) error {
	cfg, err := Parse(args)
	if errors.Is(err, lessflags.ErrHelp) {
		return nil
	}
	if err != nil {
		return err
	}
	return Serve(os.Stdin, os.Stdout, os.Stderr, cfg)
}

func Parse(args []string) (Config, error) {
	cfg, err := configFromEnv()
	if err != nil {
		return Config{}, err
	}

	var (
		name              *string
		delay, minD, maxD *time.Duration
		hang, fail        *bool
	)
	remain, err := lessflags.String("--name", &name).
		Duration("--delay", &delay).
		Duration("--delay-min", &minD).
		Duration("--delay-max", &maxD).
		Bool("--hang", &hang).
		Bool("--fail", &fail).
		Help("-h,--help", help).
		HelpNoExit().
		Parse(args)
	if err != nil {
		return Config{}, err
	}
	if len(remain) > 0 {
		return Config{}, fmt.Errorf("unexpected args: %s", strings.Join(remain, " "))
	}

	if name != nil {
		cfg.Name = strings.TrimSpace(*name)
	}
	if hang != nil {
		cfg.Hang = *hang
	}
	if fail != nil {
		cfg.Fail = *fail
	}
	if delay != nil {
		if err := checkNonNegative("--delay", *delay); err != nil {
			return Config{}, err
		}
		cfg.Delay = delay
		cfg.DelayMin = nil
		cfg.DelayMax = nil
	}
	if minD != nil {
		if err := checkNonNegative("--delay-min", *minD); err != nil {
			return Config{}, err
		}
		cfg.DelayMin = minD
		cfg.Delay = nil
	}
	if maxD != nil {
		if err := checkNonNegative("--delay-max", *maxD); err != nil {
			return Config{}, err
		}
		cfg.DelayMax = maxD
		cfg.Delay = nil
	}
	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	if cfg.Name == "" {
		cfg.Name = defaultName
	}
	return cfg, nil
}

func configFromEnv() (Config, error) {
	cfg := Config{
		Name: strings.TrimSpace(os.Getenv(EnvName)),
		Hang: envBool(EnvHang),
		Fail: envBool(EnvFail),
	}
	d, err := envDuration(EnvDelay)
	if err != nil {
		return Config{}, err
	}
	minD, err := envDuration(EnvDelayMin)
	if err != nil {
		return Config{}, err
	}
	maxD, err := envDuration(EnvDelayMax)
	if err != nil {
		return Config{}, err
	}
	if d != nil && (minD != nil || maxD != nil) {
		return Config{}, fmt.Errorf("%s cannot be set with %s/%s", EnvDelay, EnvDelayMin, EnvDelayMax)
	}
	cfg.Delay = d
	cfg.DelayMin = minD
	cfg.DelayMax = maxD
	return cfg, nil
}

func (c Config) validate() error {
	if c.Hang && c.Fail {
		return fmt.Errorf("--hang and --fail are mutually exclusive")
	}
	if c.Hang && (c.Delay != nil || c.DelayMin != nil || c.DelayMax != nil) {
		return fmt.Errorf("--hang cannot be combined with delay flags")
	}
	if c.Delay != nil && (c.DelayMin != nil || c.DelayMax != nil) {
		return fmt.Errorf("--delay cannot be combined with --delay-min/--delay-max")
	}
	if (c.DelayMin == nil) != (c.DelayMax == nil) {
		return fmt.Errorf("--delay-min and --delay-max must both be set")
	}
	if c.DelayMin != nil && c.DelayMax != nil && *c.DelayMin > *c.DelayMax {
		return fmt.Errorf("delay-min (%s) > delay-max (%s)", *c.DelayMin, *c.DelayMax)
	}
	return nil
}

func (c Config) chosenDelay() time.Duration {
	if c.Hang {
		return 0
	}
	if c.DelayMin != nil && c.DelayMax != nil {
		return pickDelay(*c.DelayMin, *c.DelayMax)
	}
	if c.Delay != nil {
		return *c.Delay
	}
	return 0
}

func checkNonNegative(name string, d time.Duration) error {
	if d < 0 {
		return fmt.Errorf("%s must be >= 0", name)
	}
	return nil
}

func envBool(key string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func envDuration(key string) (*time.Duration, error) {
	s := strings.TrimSpace(os.Getenv(key))
	if s == "" {
		return nil, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", key, err)
	}
	if d < 0 {
		return nil, fmt.Errorf("%s must be >= 0", key)
	}
	return &d, nil
}
