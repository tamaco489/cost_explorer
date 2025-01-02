package usecase

import (
	"context"
	"math"
	"strconv"
	"time"

	"github.com/tamaco489/cost_explorer/batch/internal/configuration"
	"github.com/tamaco489/cost_explorer/batch/internal/library/exchange_rates"
	"github.com/tamaco489/cost_explorer/batch/internal/library/timex"

	cost_explorer "github.com/aws/aws-sdk-go-v2/service/costexplorer"
)

type Jobber interface {
	DailyCostReport(ctx context.Context) error
	WeeklyCostReport(ctx context.Context) error
}

var _ Jobber = (*Job)(nil)

type Job struct {
	execTimeJST         time.Time
	costExplorerClient  *cost_explorer.Client
	exchangeRatesClient *exchange_rates.ExchangeRatesClient
}

func NewJob(cfg configuration.Config) (*Job, error) {

	// time
	execTimeJST := time.Now().In(timex.JST())

	// cost explorer sdk
	costExplorerClient := cost_explorer.NewFromConfig(cfg.AWSConfig)

	// open exchange rates api client
	exchangeRatesClient, err := exchange_rates.NewExchangeClient()
	if err != nil {
		return nil, err
	}

	return &Job{
		execTimeJST:         execTimeJST,
		costExplorerClient:  costExplorerClient,
		exchangeRatesClient: exchangeRatesClient,
	}, nil
}

// FIXME: 競合するので一旦退避。より適切な場所で定義する。
// parseCost: 文字列を float64 に変換する
func (j *Job) parseCost(cost string) (float64, error) {
	return strconv.ParseFloat(cost, 64)
}

// roundUpToTwoDecimalPlaces: float64 の値を小数点以下2桁で切り上げる。
func roundUpToTwoDecimalPlaces(value float64) float64 {
	factor := math.Pow(10, 2)
	return math.Ceil(value*factor) / factor
}
