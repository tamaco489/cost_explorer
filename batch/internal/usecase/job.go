package usecase

import (
	"context"
	"math"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/tamaco489/cost_explorer/batch/internal/configuration"
	"github.com/tamaco489/cost_explorer/batch/internal/library/exchange_rates"
)

type Jobber interface {
	DailyCostReport(ctx context.Context) error
	WeeklyCostReport(ctx context.Context) error
}

var _ Jobber = (*Job)(nil)

type Job struct {
	execTime            time.Time
	costExplorerClient  *costexplorer.Client
	exchangeRatesClient *exchange_rates.ExchangeRatesClient
}

func NewJob(cfg configuration.Config) (*Job, error) {

	execTime := time.Now()
	costExplorerClient := costexplorer.NewFromConfig(cfg.AWSConfig)
	exchangeRatesClient, err := exchange_rates.NewExchangeClient() // NOTE: ここでAPP_IDを指定
	if err != nil {
		return nil, err
	}

	return &Job{
		execTime:            execTime,
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
