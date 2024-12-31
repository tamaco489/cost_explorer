package slack

import "github.com/slack-go/slack"

type Attachment slack.Attachment

type slackClient struct {
	webhookURL string
	userName   string
}

// NewSlackClient: slackにメッセージを送信するためのインスタンスを初期化
func NewSlackClient(webhookURL string, userName string) *slackClient {
	return &slackClient{
		webhookURL: webhookURL,
		userName:   userName,
	}
}
