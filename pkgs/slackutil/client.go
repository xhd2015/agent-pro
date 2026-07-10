package slackutil

import (
	"os"

	"github.com/slack-go/slack"
)

// NewAPIClient creates a Slack API client honoring the SLACK_API_URL test hook.
func NewAPIClient(token string, extra ...slack.Option) *slack.Client {
	opts := append([]slack.Option{}, extra...)
	if apiURL := os.Getenv("SLACK_API_URL"); apiURL != "" {
		opts = append(opts, slack.OptionAPIURL(apiURL))
	}
	return slack.New(token, opts...)
}