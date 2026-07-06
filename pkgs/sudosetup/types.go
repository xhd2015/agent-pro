package sudosetup

// Config identifies one sudoers installation namespace.
type Config struct {
	CacheDirName string
	SudoersName  string
	Username     string // empty → resolve from env/user.Current()
}

// Rule describes the privileged command granted NOPASSWD access.
type Rule struct {
	Command     string
	ArgsPattern string // e.g. "run -c *"; empty for bare command
}

// Status is the result of Detect().
type Status struct {
	Installed, CacheWarm, CanRunNonInteractive bool
	InstallDetail, CanRunDetail, Verdict string
}

// Manifest records a successful install for persistent detection.
type Manifest struct {
	Username    string `json:"username"`
	Command     string `json:"command"`
	ArgsPattern string `json:"args_pattern"`
}