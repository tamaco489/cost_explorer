package slack

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
)

// Attachment はslackのAttachment型をラップした型です。
//
// メッセージに添付する情報を定義するために使用されます。
type Attachment slack.Attachment

// SlackClientInterface は、Slackにメッセージを送信するためのインターフェースを定義します。
//
// GoMockを使用してテスト時にモックを作成するためにこのインターフェースを定義します。
type SlackClientInterface interface {
	// SendMessage は、Slackにメッセージを送信するメソッドです。
	// 引数として、コンテキスト、メッセージのタイトル、添付ファイル（Attachment）を受け取ります。
	SendMessage(ctx context.Context, title string, attachment Attachment) error
}

// slackClient は、Slackにメッセージを送信するための構造体です。
//
// webhookURL とユーザー名を保持し、Slackへの接続に必要な情報を持ちます。
type slackClient struct {
	webhookURL string
	userName   string
}

// NewSlackClient は、Slackクライアントのインスタンスを初期化する関数です。
//
// webhookURL と userName を引数として受け取り、それらを基に slackClient を返します。
func NewSlackClient(webhookURL string, userName string) *slackClient {
	return &slackClient{
		webhookURL: webhookURL,
		userName:   userName,
	}
}

// SendMessage は、Slackにメッセージを送信するメソッドです。
//
// Webhook URL とユーザー名を使って、指定されたタイトルと添付ファイルを含むメッセージを送信します。
func (sc *slackClient) SendMessage(ctx context.Context, title string, attachment Attachment) error {
	if err := slack.PostWebhookContext(ctx, sc.webhookURL, &slack.WebhookMessage{
		Username: sc.userName,
		Text:     title,
		Attachments: []slack.Attachment{
			slack.Attachment(attachment),
		},
	}); err != nil {
		return fmt.Errorf("failed to send slack message: %w", err)
	}
	return nil
}

// ReportTitle は、レポートのタイトルを表す文字列型です。
//
// daily-report や weekly-report のような異なるレポートのタイトルを管理するために使用されます。
type ReportTitle string

// レポートタイトルとして使用する定数です。
const (
	// DailyReportTitle は、日次レポートのタイトルを表します。
	DailyReportTitle ReportTitle = "daily-cost-report"

	// WeeklyReportTitle は、週次レポートのタイトルを表します。
	WeeklyReportTitle ReportTitle = "weekly-cost-report"
)

// String: レポートタイトル型を文字列型に変換
func (rt ReportTitle) String() string {
	return string(rt)
}
