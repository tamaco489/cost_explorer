package usecase

import (
	"context"
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
