package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	service_mock "github.com/tamaco489/cost_explorer/batch/internal/service/mock"
)

func TestGetLastWeekCost(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// MockのSlackClientを生成
	mockClient := service_mock.NewMockIWeeklyCostExplorerClient(ctrl)

	// GetLastWeekCost を実行するための引数を定義
	execTime := time.Now()
	lastWeekStartDate := execTime.AddDate(0, 0, -1).Format("2006-01-02")
	lastWeekEndDate := execTime.Format("2006-01-02")

	// GetLastWeekCost を実行した結果得られるレスポンスを定義
	var expectedResponse float64 = 100

	mockClient.EXPECT().
		GetLastWeekCost(ctx, lastWeekStartDate, lastWeekEndDate).
		Return(expectedResponse, nil).
		Times(1)

	response, err := mockClient.GetLastWeekCost(ctx, lastWeekStartDate, lastWeekEndDate)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, response)
}
