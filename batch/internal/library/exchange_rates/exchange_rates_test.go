package exchange_rates_test

import (
	"context"
	"os"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
	"github.com/tamaco489/cost_explorer/batch/internal/configuration"
	"github.com/tamaco489/cost_explorer/batch/internal/library/exchange_rates"
	"github.com/tamaco489/cost_explorer/batch/internal/utils"

	exchange_rates_mock "github.com/tamaco489/cost_explorer/batch/internal/library/exchange_rates/mock"
)

func TestNewExchangeClient(t *testing.T) {
	t.Run("正常系: APP_IDが設定されている場合はインスタンスの生成に成功すること", func(t *testing.T) {
		ctx := context.Background()

		utils.PrepareTestEnvironment(ctx)
		defer os.Unsetenv("ENV")

		_, err := configuration.Load(ctx)
		assert.NoError(t, err)

		client, err := exchange_rates.NewExchangeClient()
		assert.NoError(t, err)
		assert.NotNil(t, client)

		assert.Equal(t, "test_app_id", client.AppID)
		assert.NotNil(t, client.HTTPClient)
		assert.Equal(t, "https://openexchangerates.org/api", client.BaseURL)
		assert.NotNil(t, client.BaseCurrencyFn)
	})
}

func TestPrepareExchangeRates(t *testing.T) {
	t.Run("正常系: 為替レートの準備が成功すること", func(t *testing.T) {
		client := exchange_rates.ExchangeRatesClient{
			BaseCurrencyFn: exchange_rates.GetBaseCurrency,
		}

		result, err := client.PrepareExchangeRates()
		assert.NoError(t, err)

		// 結果が期待通りであることを確認
		assert.Equal(t, exchange_rates.GetBaseCurrency(), result.BaseCurrencyCode)
		assert.Equal(t, []string{exchange_rates.JPY.String()}, result.ExchangeCurrencyCodes)
	})

	t.Run("異常系: 無効な基軸通貨が指定された場合にエラーが発生すること", func(t *testing.T) {
		client := exchange_rates.ExchangeRatesClient{
			BaseCurrencyFn: func() string {
				return "INVALID"
			},
		}

		_, err := client.PrepareExchangeRates()

		// エラーが返され、内容が期待通りであることを確認
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid base currency: INVALID")
	})
}

func TestGetExchangeRates(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// MockのSlackClientを生成
	mockClient := exchange_rates_mock.NewMockIExchangeRatesClient(ctrl)

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
