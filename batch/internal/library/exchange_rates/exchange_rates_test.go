package exchange_rates_test

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
	"github.com/tamaco489/cost_explorer/batch/internal/library/exchange_rates"

	exchange_rates_mock "github.com/tamaco489/cost_explorer/batch/internal/library/exchange_rates/mock"
)

func TestGetExchangeRates(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// MockのSlackClientを生成
	mockClient := exchange_rates_mock.NewMockExchangeRatesClientInterface(ctrl)

	// GetExchangeRates を実行するための引数を定義
	baseCurrency := exchange_rates.USD.String()
	exchangeCurrencies := []string{exchange_rates.JPY.String()}

	// GetExchangeRates を実行した結果得られるレスポンスを定義
	expectedResponse := &exchange_rates.ExchangeRatesResponse{
		Base: baseCurrency,
		Rates: map[string]float64{
			"JPY": 130.5,
		},
	}

	// モックで期待される振る舞いを設定
	mockClient.EXPECT().
		GetExchangeRates(ctx, baseCurrency, exchangeCurrencies).
		Return(expectedResponse, nil).
		Times(1) // 1回呼び出されることを期待

	// GetExchangeRates をMockで実行 (外部通信は実行されない)
	response, err := mockClient.GetExchangeRates(ctx, baseCurrency, exchangeCurrencies)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, response)
}
