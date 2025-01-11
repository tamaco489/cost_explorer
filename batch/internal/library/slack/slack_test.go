package slack_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tamaco489/cost_explorer/batch/internal/library/slack"
	"go.uber.org/mock/gomock"

	slack_mock "github.com/tamaco489/cost_explorer/batch/internal/library/slack/mock"
)

// TestSendMessage: モックを使用して slackClient の SendMessage メソッドをテストします
func TestSendMessage(t *testing.T) {
	ctx := context.Background()
	messageTitle := slack.DailyReportTitle.String()
	var sa slack.Attachment

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// モックのSlackClientを生成
	mockSlackClient := slack_mock.NewMockSlackClientInterface(ctrl)

	// モックで期待される振る舞いを設定
	mockSlackClient.EXPECT().
		SendMessage(ctx, messageTitle, sa).
		Return(nil).
		Times(1) // 1回呼び出されることを期待

	// SendMessage をモックしたものを使って呼び出し (外部通信は行われない)
	err := mockSlackClient.SendMessage(ctx, messageTitle, sa)
	assert.NoError(t, err)
}
