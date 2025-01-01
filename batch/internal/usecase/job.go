package usecase

import (
	"context"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/tamaco489/cost_explorer/batch/internal/configuration"
)

type Jobber interface {
	DailyCostReport(ctx context.Context) error
	WeeklyCostReport(ctx context.Context) error
}

var _ Jobber = (*Job)(nil)

type Job struct {
	execTime           time.Time
	costExplorerClient *costexplorer.Client
}

func NewJob(cfg configuration.Config) (*Job, error) {

	execTime := time.Now()
	costExplorerClient := costexplorer.NewFromConfig(cfg.AWSConfig)

	return &Job{
		execTime:           execTime,
		costExplorerClient: costExplorerClient,
	}, nil
}

// FIXME: 競合するので一旦退避。より適切な場所で定義する。
// parseCost: 文字列を float64 に変換する
func (j *Job) parseCost(cost string) (float64, error) {
	return strconv.ParseFloat(cost, 64)
}
