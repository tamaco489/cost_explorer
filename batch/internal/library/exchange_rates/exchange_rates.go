package exchange_rates

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/tamaco489/cost_explorer/batch/internal/configuration"
)

const baseURL string = "https://openexchangerates.org/api"

type IExchangeRatesClient interface {
	GetExchangeRates(ctx context.Context, baseCurrencyCode string, exchangeCurrencyCodes []string) (*ExchangeRatesResponse, error)
}

var _ IExchangeRatesClient = (*ExchangeRatesClient)(nil)

type ExchangeRatesClient struct {
	AppID          string
	HTTPClient     *http.Client
	BaseURL        string
	BaseCurrencyFn func() string // 基軸通貨を取得する関数
}

type ExchangeRatesResponse struct {
	Disclaimer string             `json:"disclaimer"`
	License    string             `json:"license"`
	Timestamp  int64              `json:"timestamp"`
	Base       string             `json:"base"`
	Rates      map[string]float64 `json:"rates"`
}

// NewExchangeClient: GetExchangeRates のコンストラクタ
func NewExchangeClient() (*ExchangeRatesClient, error) {

	appID := configuration.Get().ExchangeRates.AppID
	if appID == "" {
		return nil, errors.New("APP_ID is required")
	}

	client := &ExchangeRatesClient{
		AppID:          appID,
		HTTPClient:     &http.Client{Timeout: 10 * time.Second},
		BaseURL:        baseURL,
		BaseCurrencyFn: GetBaseCurrency,
	}

	return client, nil
}

// PrepareExchangeRates: GetExchangeRates を実行するにあたっての準備
//
// 基軸通貨をUSDに設定し、変換対象通貨をJPYに限定
func (erc *ExchangeRatesClient) PrepareExchangeRates() (*prepareExchangeRates, error) {
	baseCurrencyCode := erc.BaseCurrencyFn()
	if !ExchangeRatesCurrencyCode(baseCurrencyCode).Valid() {
		return nil, fmt.Errorf("invalid base currency: %s", baseCurrencyCode)
	}

	return &prepareExchangeRates{
		BaseCurrencyCode:      baseCurrencyCode,
		ExchangeCurrencyCodes: []string{JPY.String()},
	}, nil
}

type prepareExchangeRates struct {
	BaseCurrencyCode      string
	ExchangeCurrencyCodes []string
}

// GetExchangeRates: 為替レートを取得
func (erc *ExchangeRatesClient) GetExchangeRates(ctx context.Context, baseCurrencyCode string, exchangeCurrencyCodes []string) (*ExchangeRatesResponse, error) {

	symbolsParam := strings.Join(exchangeCurrencyCodes, ",")

	url := fmt.Sprintf("%s/latest.json?app_id=%s&base=%s&symbols=%s", erc.BaseURL, erc.AppID, baseCurrencyCode, symbolsParam)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := erc.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get exchange rates, status code: %d", resp.StatusCode)
	}

	var ratesResponse ExchangeRatesResponse
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&ratesResponse); err != nil {
		return nil, err
	}

	return &ratesResponse, nil
}
