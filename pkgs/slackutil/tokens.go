package slackutil

import "os"

// ResolveBotToken returns the bot token from CLI flag, env, then config.
func ResolveBotToken(flagValue string, envKey string, cfg *SlackConfig) string {
	if flagValue != "" {
		return flagValue
	}
	if env := os.Getenv(envKey); env != "" {
		return env
	}
	if cfg != nil {
		return cfg.BotToken
	}
	return ""
}

// ResolveAppToken returns the app token from CLI flag, env, then config.
func ResolveAppToken(flagValue string, envKey string, cfg *SlackConfig) string {
	if flagValue != "" {
		return flagValue
	}
	if env := os.Getenv(envKey); env != "" {
		return env
	}
	if cfg != nil {
		return cfg.AppToken
	}
	return ""
}