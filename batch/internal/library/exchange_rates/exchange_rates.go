package exchange_rates

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/tamaco489/cost_explorer/batch/internal/configuration"
)

const baseURL string = "https://openexchangerates.org/api"

type ExchangeRatesClient struct {
	AppID      string
	HTTPClient *http.Client
	BaseURL    string
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
		AppID:      appID,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		BaseURL:    baseURL,
	}

	return client, nil
}

// GetExchangeRates: 為替レートを取得する
func (erc *ExchangeRatesClient) GetExchangeRates(baseCurrencyCode string, exchangeCurrencyCodes []string) (*ExchangeRatesResponse, error) {

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
