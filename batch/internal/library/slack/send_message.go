package slack

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
)

// レポートタイトルの型を定義
type ReportTitle string

const (
	DailyCostReportTitle  ReportTitle = "daily-cost-report"
	WeeklyCostReportTitle ReportTitle = "weekly-cost-report"
)

// String: レポートタイトル型を文字列型に変換する
func (rt ReportTitle) String() string {
	return string(rt)
}

// SendMessage: slackにメッセージを送信する
func (sc *slackClient) SendMessage(ctx context.Context, title string, attachment Attachment) error {
	err := slack.PostWebhookContext(ctx, sc.webhookURL, &slack.WebhookMessage{
		Username: sc.userName,
		Text:     title,
		Attachments: []slack.Attachment{
			slack.Attachment(attachment),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to send slack message: %w", err)
	}
	return nil
}
