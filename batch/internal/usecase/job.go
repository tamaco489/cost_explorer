package usecase

import (
	"context"
	"time"

	"github.com/tamaco489/cost_explorer/batch/internal/configuration"
	"github.com/tamaco489/cost_explorer/batch/internal/library/exchange_rates"
	"github.com/tamaco489/cost_explorer/batch/internal/library/timex"
	"github.com/tamaco489/cost_explorer/batch/internal/service"

	cost_explorer "github.com/aws/aws-sdk-go-v2/service/costexplorer"
)

type Jobber interface {
	DailyCostReport(ctx context.Context) error
	WeeklyCostReport(ctx context.Context) error
}

var _ Jobber = (*Job)(nil)

type Job struct {
	execTimeJST               time.Time
	dailyCostExplorerService  *service.DailyCostExplorerService
	weeklyCostExplorerService *service.WeeklyCostExplorerService
	exchangeRatesClient       *exchange_rates.ExchangeRatesClient
}

func NewJob(cfg configuration.Config) (*Job, error) {

	// time
	execTimeJST := time.Now().In(timex.JST())

	// cost explorer sdk
	costExplorerClient := cost_explorer.NewFromConfig(cfg.AWSConfig)
	dailyCostExplorerService := service.NewDailyCostExplorerService(costExplorerClient)
	weeklyCostExplorerService := service.NewWeeklyCostExplorerService(costExplorerClient)

	// open exchange rates api client
	exchangeRatesClient, err := exchange_rates.NewExchangeClient()
	if err != nil {
		return nil, err
	}

	return &Job{
		execTimeJST:               execTimeJST,
		dailyCostExplorerService:  dailyCostExplorerService,
		weeklyCostExplorerService: weeklyCostExplorerService,
		exchangeRatesClient:       exchangeRatesClient,
	}, nil
}
