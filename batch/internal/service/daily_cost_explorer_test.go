package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	service_mock "github.com/tamaco489/cost_explorer/batch/internal/service/mock"
)

func TestGetYesterdayCost(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// MockのSlackClientを生成
	mockClient := service_mock.NewMockICostExplorerClient(ctrl)

	// GetYesterdayCost を実行するための引数を定義
	execTime := time.Now()
	yesterday := execTime.AddDate(0, 0, -1).Format("2006-01-02")
	endDate := execTime.Format("2006-01-02")

	// GetYesterdayCost を実行した結果得られるレスポンスを定義
	var expectedResponse float64 = 100

	mockClient.EXPECT().
		GetYesterdayCost(ctx, yesterday, endDate).
		Return(expectedResponse, nil).
		Times(1)

	response, err := mockClient.GetYesterdayCost(ctx, yesterday, endDate)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, response)
}
